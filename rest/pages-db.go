package rest

import (
  "database/sql"
  "fmt"
  "net/http"
  "strconv"

  "google.golang.org/appengine"
)

type ResultBuilder func(*sql.Rows) (interface{}, error)

func ProcessPageStmt(queryBase string, whereBase string, searchParams *SearchParams, params []interface{}) (string, []interface{}, RestError) {
  if len(searchParams.Scopes) > 1 {
    return "", nil, BadRequestError("We currently only support a single scope.", nil)
  } else if (len(searchParams.Scopes) == 0) {
    return "", nil, BadRequestError("No scope specified.", nil)
  }

  customJoin := ""
  customWhere := ""

  if searchParams.Scopes[0] == "Active" {
    customJoin = "JOIN packages p ON p.customer_id=c.id AND p.status IN ('CREATED', 'REJECTED', 'ACCEPTED', 'PACKAGED', 'PICKED_UP', 'SORTED', 'OUT_FOR_DELIVERY') "
  } else if searchParams.Scopes[0] != "All" {
    return "", nil, BadRequestError(fmt.Sprintf("Found unknown scope: '%s'.", searchParams.Scopes[0]), nil)
  }

  for _, term := range searchParams.Terms {
    likeTerm := "%"+term+"%"
    if _, err := strconv.ParseInt(term,10,64); err == nil {
      customWhere += "AND (c.phone LIKE ? OR c.phone_backup LIKE ?) "
      params = append(params, likeTerm, likeTerm)
    } else {
      customWhere += "AND (c.name LIKE ? OR c.email LIKE ?) "
      params = append(params, likeTerm, likeTerm)
    }
  }

  // LIMIT
  // 1-based index from sturct; need 0-based here; itemsPerPage and pageIndex
  // are set to defaults and within constraints at the API level.
  pageIndex := searchParams.PageInfo.PageIndex - 1
  itemsPerPage := searchParams.PageInfo.ItemsPerPage

  limitAndOrderBy := `ORDER BY `

  // ORDER BY
  if searchParams.Sort == `name-asc` {
    limitAndOrderBy += `c.name ASC `
  } else if searchParams.Sort == `name-desc` {
    limitAndOrderBy += `c.name DESC `
  } else {
    return "", nil, UnprocessableEntityError(fmt.Sprintf("Bad sort value: '%s'.", searchParams.Sort), nil)
  }

  limitAndOrderBy += `LIMIT ` + strconv.Itoa(pageIndex * itemsPerPage) + `, ` + strconv.Itoa(itemsPerPage)

  queryStmt := queryBase + customJoin + whereBase + customWhere + limitAndOrderBy

  return queryStmt, params, nil
}

func ProcessListQuery(conn *sql.DB, queryStmt string, params []interface{}, resultBuilder ResultBuilder, r *http.Request) (interface{}, int64, RestError) {
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
    return nil, 0, ServerError(fmt.Sprintf("Could not create package record."), err)
  }

  query, err := txn.Prepare(queryStmt)
  if err != nil {
    txn.Rollback()
    return nil, 0, ServerError("Could not process query.", err)
  }

  rows, err := query.Query(params...)
  if err != nil {
    txn.Rollback()
    return nil, 0, ServerError("Could not retrieve customers.", err)
  }

  // This block must come before the 'SELECT FOUND_ROWS()'. My guess is it's
  // related to 'rows.Next()'.
  results, err := resultBuilder(rows)
  if err != nil {
    rows.Close()
    txn.Rollback()
    return nil, 0, ServerError("Could not retrieve customers.", err)
  }

  // Notice no 'defer'. If we don't close the row right away, then we get
  // 'busy buffer' errors when executing the next txn query.
  rows.Close()

  countRows, err := txn.Query("SELECT FOUND_ROWS()")
  if err != nil {
    txn.Rollback()
    return nil, 0, ServerError("Could not retrieve customers.", err)
  }

  var count int64
  if !countRows.Next() {
    countRows.Close()
    txn.Rollback()
    return nil, 0, ServerError("Could not retrieve customers.", err)
  }
  if err = countRows.Scan(&count); err != nil {
    countRows.Close()
    txn.Rollback()
    return nil, 0, ServerError("Could not retrieve customers.", err)
  }
  countRows.Close()
  txn.Commit()

  return results, count, nil
}
