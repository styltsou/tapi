package ui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/styltsou/tapi/internal/storage"
)

func deleteRequestCmd(collectionName, requestName string) tea.Cmd {
	return func() tea.Msg {
		collections, _ := storage.LoadCollections()
		for i, col := range collections {
			if col.Name == collectionName {
				// Filter out the request
				newRequests := []storage.Request{}
				for _, r := range col.Requests {
					if r.Name != requestName {
						newRequests = append(newRequests, r)
					}
				}
				collections[i].Requests = newRequests
				if err := storage.SaveCollection(collections[i]); err != nil {
					return ErrMsg{Err: err}
				}
				return tea.Batch(
					showStatusCmd("Request deleted", false),
					loadCollectionsCmd(),
				)
			}
		}
		return ErrMsg{Err: fmt.Errorf("collection %q not found", collectionName)}
	}
}

func deleteCollectionCmd(name string) tea.Cmd {
	return func() tea.Msg {
		if err := storage.DeleteCollection(name); err != nil {
			return ErrMsg{Err: err}
		}
		return tea.Batch(
			showStatusCmd("Collection deleted", false),
			loadCollectionsCmd(),
		)
	}
}

func renameCollectionCmd(oldName, newName string) tea.Cmd {
	return func() tea.Msg {
		collections, _ := storage.LoadCollections()
		for _, col := range collections {
			if col.Name == oldName {
				// 1. Delete old
				if err := storage.DeleteCollection(oldName); err != nil {
					return ErrMsg{Err: err}
				}
				// 2. Save as new
				col.Name = newName
				if err := storage.SaveCollection(col); err != nil {
					return ErrMsg{Err: err}
				}
				return tea.Batch(
					showStatusCmd("Collection renamed", false),
					loadCollectionsCmd(),
				)
			}
		}
		return ErrMsg{Err: fmt.Errorf("collection not found")}
	}
}

func duplicateRequestCmd(collectionName, requestName string) tea.Cmd {
	return func() tea.Msg {
		collections, _ := storage.LoadCollections()
		for i, col := range collections {
			if col.Name == collectionName {
				for _, r := range col.Requests {
					if r.Name == requestName {
						// Clone the request with a " (copy)" suffix
						dup := storage.Request{
							Name:    r.Name + " (copy)",
							Method:  r.Method,
							URL:     r.URL,
							Body:    r.Body,
							Headers: make(map[string]string),
						}
						for k, v := range r.Headers {
							dup.Headers[k] = v
						}
						collections[i].Requests = append(collections[i].Requests, dup)
						if err := storage.SaveCollection(collections[i]); err != nil {
							return ErrMsg{Err: err}
						}
						return tea.Batch(
							showStatusCmd("Request duplicated", false),
							loadCollectionsCmd(),
						)
					}
				}
			}
		}
		return ErrMsg{Err: fmt.Errorf("request not found")}
	}
}
