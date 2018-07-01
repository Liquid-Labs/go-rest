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
