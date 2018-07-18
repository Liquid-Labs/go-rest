package rest

import (
  "database/sql"
  "fmt"
  "net/http"
  "strconv"
  "strings"

  "google.golang.org/appengine"
)

type ResultBuilder func(*sql.Rows) (interface{}, error)

/**
 * Takes the raw search term and current param list. Returns WHERE-clause
 * conditions statement and updated params array.
 */
type WhereBitGenerator func(string, []interface{}) (string, []interface{}, error)

type JoinData struct {
  JoinClause  string
  WhereClause string
  JoinParams  []interface{}
}

/**
 * Struct defining all the necessary parameters to build a paged query.
 */
type PagedQueryParameters struct {
  // Type based query parameters.
  FieldSpec            string // specifies fields to select
  GeneralFrom          string // specifies structural joins; e.g., to denormalize data
  ScopeJoins           map[string]JoinData // defines possible joins based on SearchParams.Scope
  SearchWhereGenerator WhereBitGenerator // generats query and determines parameters based on SearchParams
  SortMap              map[string]string // clauses for each sorting option
  // Type based result helpers
  ResultBuilder        ResultBuilder
  ResourceName         string
  // Query specific parameters
  SearchParams         *SearchParams
  ContextJoins         []JoinData // Join data based of call context; e.g., '/store/xxx/customers'
  // Plumbing
  Db                   *sql.DB
  Request              *http.Request
}

// TODO: modifying 'SearchParams' paging info should result in warning being included in the final results.
/**
 * Executes a page-based list query, returning results with paging meta data.
 *
 * Returns:
 * 0) May modify contentns of 'SearchParams' to reflect actual paging parameters
 *    used. All other parameters should be unmodified.
 * 1) List of results of any type. Underyling type specified by the
 *    ResultBuilder.
 * 2) Total count of stuff.
 * 3) Any error.
 */
func PagedQuery(pqp PagedQueryParameters, contextJoins []*JoinData) (interface{}, int64, RestError) {
  // First, we construct the query string
  params := make([]interface{}, 0) // build up query parameters
  fromBit := pqp.GeneralFrom // build up 'FROM' clause
  whereBit := "WHERE TRUE " // build up 'WHERE clause'; include the 'TRUE' so everyone can blindly continue with 'AND'

  for _, contextJoin := range contextJoins {
    fromBit += contextJoin.JoinClause
    whereBit += contextJoin.WhereClause
    params = append(params, contextJoin.JoinParams...)
  }

  for _, scope := range pqp.SearchParams.Scopes {
    if val, ok := pqp.ScopeJoins[scope]; !ok {
      return nil, -1, BadRequestError(fmt.Sprintf("Found unknown scope: '%s'.", pqp.SearchParams.Scopes[0]), nil)
    } else {
      fromBit += val.JoinClause
      whereBit += val.WhereClause
      params = append(params, val.JoinParams...)
    }
  }

  for _, term := range pqp.SearchParams.Terms {
    var whereTerm string
    var err error
    whereTerm, params, err = pqp.SearchWhereGenerator(term, params)
    if err != nil {
      return nil, -1, BadRequestError(fmt.Sprintf("Could not process search term: '%s'.", term), err)
    } else {
      whereBit += whereTerm
    }
  }

  // LIMIT
  // 1-based index from sturct; need 0-based here; itemsPerPage and pageIndex
  // are set to defaults and within constraints at the API level.
  pageIndex := pqp.SearchParams.PageInfo.PageIndex - 1
  itemsPerPage := pqp.SearchParams.PageInfo.ItemsPerPage

  // ORDER BY
  var limitAndOrderBy string
  if pqp.SearchParams.Sort != "" {
    limitAndOrderBy = `ORDER BY `
    if val, ok := pqp.SortMap[pqp.SearchParams.Sort]; !ok {
      return nil, -1, UnprocessableEntityError(fmt.Sprintf("Bad sort value: '%s'.", pqp.SearchParams.Sort), nil)
    } else {
      limitAndOrderBy += val
    }
  } else {
    limitAndOrderBy = ""
  }

  limitAndOrderBy += `LIMIT ` + strconv.Itoa(pageIndex * itemsPerPage) + `, ` + strconv.Itoa(itemsPerPage)

  if whereBit == "WHERE TRUE " {
    whereBit = ""
  }

  // And here it is:
  queryStmt := strings.Join([]string{
    `SELECT DISTINCT SQL_CALC_FOUND_ROWS `,
    pqp.FieldSpec,
    fromBit,
    whereBit,
    limitAndOrderBy}, "")

  // If we do not run the search + total row count quries in a txn, then we
  // don't get reliable results from the total row count. I suspect it's because
  // different connections are being used and the txn forces it all onto a single
  // connection.
  //
  // Note, we also avoid 'defer' to '.Close()' the rows or '.Rollback()/Commit()'
  // the txn. It generates 'busy buffer' errors when used with a txn.
  ctx := appengine.NewContext(pqp.Request)
  txn, err := pqp.Db.BeginTx(ctx, nil)
  if err != nil {
    txn.Rollback()
    return nil, 0, ServerError(fmt.Sprintf("Could not retrieve " + pqp.ResourceName + "."), err)
  }

  query, err := txn.Prepare(queryStmt)
  if err != nil {
    txn.Rollback()
    return nil, 0, ServerError("Could not process " + pqp.ResourceName + " query.", err)
  }

  rows, err := query.Query(params...)
  if err != nil {
    txn.Rollback()
    return nil, 0, ServerError("Could not retrieve " + pqp.ResourceName + ".", err)
  }

  // This block must come before the 'SELECT FOUND_ROWS()'. My guess is it's
  // related to 'rows.Next()'.
  results, err := pqp.ResultBuilder(rows)
  if err != nil {
    rows.Close()
    txn.Rollback()
    return nil, 0, ServerError("Could not retrieve " + pqp.ResourceName + ".", err)
  }

  // Notice no 'defer'. If we don't close the row right away, then we get
  // 'busy buffer' errors when executing the next txn query.
  rows.Close()

  countRows, err := txn.Query("SELECT FOUND_ROWS()")
  if err != nil {
    txn.Rollback()
    return nil, 0, ServerError("Could not retrieve " + pqp.ResourceName + ".", err)
  }

  var count int64
  if !countRows.Next() {
    countRows.Close()
    txn.Rollback()
    return nil, 0, ServerError("Could not retrieve " + pqp.ResourceName + ".", err)
  }
  if err = countRows.Scan(&count); err != nil {
    countRows.Close()
    txn.Rollback()
    return nil, 0, ServerError("Could not retrieve " + pqp.ResourceName + ".", err)
  }
  countRows.Close()
  txn.Commit()

  return results, count, nil
 }
