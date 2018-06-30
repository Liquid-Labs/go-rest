package rest

import (
  "fmt"
  "net/http"
  "strconv"
  "strings"
)

type PageInfo struct {
  // 1-based index
  PageIndex      int   `json:"pageIndex"`
  ItemsPerPage   int   `json:"itemsPerPage"`
  TotalItemCount int64 `json:"totalItemCount"`
  TotalPageCount int64 `json:"totalPageCount"`
}

type SearchParams struct {
  Scopes   []string  `json:"scopes"`
  Terms    []string  `json:"terms"`
  Sort     string    `json:"sort"`
  PageInfo *PageInfo `json:"pageInfo"`
}

func ExtractSearchParamsFromUrl(w http.ResponseWriter, r *http.Request) (*SearchParams, RestError) {
  var sp SearchParams

  scopes := r.FormValue("scopes")
  if scopes != "" {
    sp.Scopes = strings.Split(scopes, ",")
  } else {
    sp.Scopes = make([]string, 0)
  }
  terms := r.FormValue("terms")
  if terms != "" {
    sp.Terms = strings.Split(terms, ",")
  } else {
    sp.Terms = make([]string, 0)
  }
  sp.Sort = r.FormValue("sort")

  pageIndexStr := r.FormValue("pageIndex")
  var pageIndex int
  var err error
  if pageIndexStr != "" {
    if pageIndex, err = strconv.Atoi(pageIndexStr); err != nil {
      return nil, HandleError(w, BadRequestError(fmt.Sprintf("Could not parse pageIndex: %s", pageIndexStr), err))
    }
  } else {
    pageIndex = 1
  }

  itemsPerPageStr := r.FormValue("itemsPerPage")
  var itemsPerPage int
  if itemsPerPageStr != "" {
    if itemsPerPage, err = strconv.Atoi(itemsPerPageStr); err != nil {
      return nil, HandleError(w, BadRequestError(fmt.Sprintf("Could not parse itemsPerPage: %s", itemsPerPageStr), err))
    }
    if itemsPerPage < 20 {
      itemsPerPage = 20
    } else if itemsPerPage > 250 {
      itemsPerPage = 250
    }
  } else {
    itemsPerPage = 100
  }
  sp.PageInfo = &PageInfo{PageIndex: pageIndex, ItemsPerPage: itemsPerPage}

  return &sp, nil
}

func (sp *SearchParams) SetTotalPages(count int64) {
    var itemsPerPage int = sp.PageInfo.ItemsPerPage
    var pageIndex int = sp.PageInfo.PageIndex
    pageCount := count/int64(itemsPerPage)
    if count % int64(itemsPerPage) > 0 {
      pageCount += 1
    }
    sp.PageInfo = &PageInfo{PageIndex: int(pageIndex), ItemsPerPage: int(itemsPerPage), TotalItemCount: count, TotalPageCount: pageCount}
}

func ProcessSearchParams(searchParams *SearchParams) (string, string, string, []interface{}, RestError) {
  if len(searchParams.Scopes) > 1 {
    return "", "", "", nil, BadRequestError("We currently only support a single scope.", nil)
  } else if (len(searchParams.Scopes) == 0) {
    return "", "", "", nil, BadRequestError("No scope specified.", nil)
  }

  customJoin := ""
  customWhere := ""
  params := make([]interface{}, 0)

  if searchParams.Scopes[0] == "Active" {
    customJoin = "JOIN packages p ON p.customer_id=c.id AND p.status IN ('CREATED', 'REJECTED', 'ACCEPTED', 'PACKAGED', 'PICKED_UP', 'SORTED', 'OUT_FOR_DELIVERY') "
  } else if searchParams.Scopes[0] != "All" {
    return "", "", "", nil, BadRequestError(fmt.Sprintf("Found unknown scope: '%s'.", searchParams.Scopes[0]), nil)
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
    return "", "", "", nil, UnprocessableEntityError(fmt.Sprintf("Bad sort value: '%s'.", searchParams.Sort), nil)
  }

  limitAndOrderBy += `LIMIT ` + strconv.Itoa(pageIndex * itemsPerPage) + `, ` + strconv.Itoa(itemsPerPage)

  return customJoin, customWhere, limitAndOrderBy, params, nil
}
