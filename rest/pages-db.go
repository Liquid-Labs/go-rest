package rest

import (
  "database/sql"
  "fmt"
  "net/http"
  "strconv"

  "google.golang.org/appengine"
)

type ResultBuilder func(*sql.Rows) (interface{}, error)

type GeneralSearchWhereBit func(string, string, []interface{}) (string, []interface{}, error)

func ProcessPageStmt(queryBase string, whereBase string, scopeJoins map[string]string, whereFunc GeneralSearchWhereBit, sortMap map[string]string, searchParams *SearchParams, params []interface{}) (string, []interface{}, RestError) {
  if err := searchParams.EnsureSingleScope(); err != nil {
    return "", nil, err
  }

  customJoin := ""
  customWhere := ""

  if val, ok := scopeJoins[searchParams.Scopes[0]]; !ok {
    return "", nil, BadRequestError(fmt.Sprintf("Found unknown scope: '%s'.", searchParams.Scopes[0]), nil)
  } else {
    customJoin = val;
  }

  for _, term := range searchParams.Terms {
    likeTerm := "%"+term+"%"

    var whereBit string
    var err error
    whereBit, params, err = whereFunc(term, likeTerm, params)
    if err != nil {
      return "", nil, BadRequestError(fmt.Sprintf("Could not process search term: '%s'.", term), err)
    } else {
      customWhere += whereBit
    }
  }

  // LIMIT
  // 1-based index from sturct; need 0-based here; itemsPerPage and pageIndex
  // are set to defaults and within constraints at the API level.
  pageIndex := searchParams.PageInfo.PageIndex - 1
  itemsPerPage := searchParams.PageInfo.ItemsPerPage

  // ORDER BY
  limitAndOrderBy := `ORDER BY `
  if val, ok := sortMap[searchParams.Sort]; !ok {
    return "", nil, UnprocessableEntityError(fmt.Sprintf("Bad sort value: '%s'.", searchParams.Sort), nil)
  } else {
    limitAndOrderBy += val
  }

  limitAndOrderBy += `LIMIT ` + strconv.Itoa(pageIndex * itemsPerPage) + `, ` + strconv.Itoa(itemsPerPage)

  queryStmt := queryBase + customJoin + whereBase + customWhere + limitAndOrderBy

  return queryStmt, params, nil
}

func ProcessListQuery(conn *sql.DB, queryStmt string, params []interface{}, resultBuilder ResultBuilder, r *http.Request, resourceName string) (interface{}, int64, RestError) {
  // If we do not run the search + total row count quries in a txn, then we
  // don't get reliable results from the total row count. I suspect it's because
  // different connections are being used and the txn forces it all onto a single
  // connection.
  //
  // Note, we also avoid 'defer' to '.Close()' the rows or '.Rollback()/Commit()'
  // the txn. It generates 'busy buffer' errors when used with a txn.
  ctx := appengine.NewContext(r)
  txn, err := conn.BeginTx(ctx, nil)
  if err != nil {
    txn.Rollback()
    return nil, 0, ServerError(fmt.Sprintf("Could not retrieve " + resourceName + "."), err)
  }

  query, err := txn.Prepare(queryStmt)
  if err != nil {
    txn.Rollback()
    return nil, 0, ServerError("Could not process " + resourceName + " query.", err)
  }

  rows, err := query.Query(params...)
  if err != nil {
    txn.Rollback()
    return nil, 0, ServerError("Could not retrieve " + resourceName + ".", err)
  }

  // This block must come before the 'SELECT FOUND_ROWS()'. My guess is it's
  // related to 'rows.Next()'.
  results, err := resultBuilder(rows)
  if err != nil {
    rows.Close()
    txn.Rollback()
    return nil, 0, ServerError("Could not retrieve " + resourceName + ".", err)
  }

  // Notice no 'defer'. If we don't close the row right away, then we get
  // 'busy buffer' errors when executing the next txn query.
  rows.Close()

  countRows, err := txn.Query("SELECT FOUND_ROWS()")
  if err != nil {
    txn.Rollback()
    return nil, 0, ServerError("Could not retrieve " + resourceName + ".", err)
  }

  var count int64
  if !countRows.Next() {
    countRows.Close()
    txn.Rollback()
    return nil, 0, ServerError("Could not retrieve " + resourceName + ".", err)
  }
  if err = countRows.Scan(&count); err != nil {
    countRows.Close()
    txn.Rollback()
    return nil, 0, ServerError("Could not retrieve " + resourceName + ".", err)
  }
  countRows.Close()
  txn.Commit()

  return results, count, nil
}
