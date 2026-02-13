package commands

import (
	uimsg "github.com/styltsou/tapi/internal/ui/msg"

	"fmt"

	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/styltsou/tapi/internal/http"
	"github.com/styltsou/tapi/internal/storage"
)

func DeleteRequestCmd(collectionName, requestName string) tea.Cmd {
	return func() tea.Msg {
		collections, _, _ := storage.LoadCollections()
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
					return uimsg.ErrMsg{Err: err}
				}
				return uimsg.RequestDeletedMsg{
					CollectionName: collectionName,
					RequestName:    requestName,
				}
			}
		}
		return uimsg.ErrMsg{Err: fmt.Errorf("collection %q not found", collectionName)}
	}
}

func DeleteCollectionCmd(name string) tea.Cmd {
	return func() tea.Msg {
		if err := storage.DeleteCollection(name); err != nil {
			return uimsg.ErrMsg{Err: err}
		}
		return uimsg.CollectionDeletedMsg{Name: name}
	}
}

func RenameCollectionCmd(oldName, newName string) tea.Cmd {
	return func() tea.Msg {
		collections, _, _ := storage.LoadCollections()
		for _, col := range collections {
			if col.Name == oldName {
				// 1. Delete old
				if err := storage.DeleteCollection(oldName); err != nil {
					return uimsg.ErrMsg{Err: err}
				}
				// 2. Save as new
				col.Name = newName
				if err := storage.SaveCollection(col); err != nil {
					return uimsg.ErrMsg{Err: err}
				}
				return uimsg.CollectionRenamedMsg{OldName: oldName, NewName: newName}
			}
		}
		return uimsg.ErrMsg{Err: fmt.Errorf("collection not found")}
	}
}

func DuplicateRequestCmd(collectionName, requestName string) tea.Cmd {
	return func() tea.Msg {
		collections, _, _ := storage.LoadCollections()
		for i, col := range collections {
			if col.Name == collectionName {
				for _, r := range col.Requests {
					if r.Name == requestName {
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
						if r.Auth != nil {
							dup.Auth = &storage.BasicAuth{
								Username: r.Auth.Username,
								Password: r.Auth.Password,
							}
						}
						collections[i].Requests = append(collections[i].Requests, dup)
						if err := storage.SaveCollection(collections[i]); err != nil {
							return uimsg.ErrMsg{Err: err}
						}
						return uimsg.RequestDuplicatedMsg{
							CollectionName: collectionName,
							RequestName:    r.Name + " (copy)",
						}
					}
				}
			}
		}
		return uimsg.ErrMsg{Err: fmt.Errorf("request not found")}
	}
}

func LoadCollectionsCmd() tea.Cmd {
	return func() tea.Msg {
		collections, loadErrs, err := storage.LoadCollections()
		if err != nil {
			return uimsg.ErrMsg{Err: err}
		}
		return uimsg.CollectionsLoadedMsg{
			Collections: collections,
			LoadErrors:  loadErrs,
		}
	}
}

func ExecuteRequestCmd(httpClient *http.Client, req storage.Request, baseURL string) tea.Cmd {
	return func() tea.Msg {
		response, err := httpClient.Execute(req, baseURL)
		if err != nil {
			return uimsg.ErrMsg{Err: err}
		}
		return uimsg.ResponseReadyMsg{Response: response, Request: req}
	}
}

func SaveRequestCmd(req storage.Request) tea.Cmd {
	return func() tea.Msg {
		collections, _, _ := storage.LoadCollections()
		for i, col := range collections {
			for j, r := range col.Requests {
				if r.Name == req.Name {
					collections[i].Requests[j] = req
					if err := storage.SaveCollection(collections[i]); err != nil {
						return uimsg.ErrMsg{Err: err}
					}
					return uimsg.StatusMsg{Message: "Request saved", IsError: false}
				}
			}
		}
		return uimsg.ErrMsg{Err: fmt.Errorf("collection not found for request %s", req.Name)}
	}
}

func SaveCollectionCmd(col storage.Collection) tea.Cmd {
	return func() tea.Msg {
		if err := storage.SaveCollection(col); err != nil {
			return uimsg.ErrMsg{Err: err}
		}
		return uimsg.StatusMsg{Message: "Collection saved", IsError: false}
	}
}

func ShowStatusCmd(message string, isError bool) tea.Cmd {
	return func() tea.Msg {
		return uimsg.StatusMsg{Message: message, IsError: isError}
	}
}

func LoadEnvsCmd() tea.Cmd {
	return func() tea.Msg {
		envs, err := storage.LoadEnvironments()
		if err != nil {
			return uimsg.ErrMsg{Err: err}
		}
		return uimsg.EnvsLoadedMsg{Envs: envs}
	}
}

func SaveEnvCmd(env storage.Environment) tea.Cmd {
	return func() tea.Msg {
		if err := storage.SaveEnvironment(env); err != nil {
			return uimsg.ErrMsg{Err: err}
		}
		return uimsg.EnvSavedMsg{Env: env}
	}
}

func DeleteEnvCmd(name string) tea.Cmd {
	return func() tea.Msg {
		if err := storage.DeleteEnvironment(name); err != nil {
			return uimsg.ErrMsg{Err: err}
		}
		return uimsg.EnvDeletedMsg{Name: name}
	}
}

func CreateRequestCmd(collectionName string, req storage.Request) tea.Cmd {
	return func() tea.Msg {
		collections, _, _ := storage.LoadCollections()
		
		// Find or create collection
		var targetCol *storage.Collection
		var targetIdx int
		found := false

		for i := range collections {
			if collections[i].Name == collectionName {
				targetCol = &collections[i]
				targetIdx = i
				found = true
				break
			}
		}

		if !found {
			// Create new collection
			newCol := storage.Collection{Name: collectionName, Requests: []storage.Request{req}}
			if err := storage.SaveCollection(newCol); err != nil {
				return uimsg.ErrMsg{Err: err}
			}
		} else {
			// Add to existing
			targetCol.Requests = append(targetCol.Requests, req)
			// Update in list
			collections[targetIdx] = *targetCol
			if err := storage.SaveCollection(*targetCol); err != nil {
				return uimsg.ErrMsg{Err: err}
			}
		}

		return uimsg.RequestCreatedMsg{
			CollectionName: collectionName,
			Request:        req,
		}
	}
}

func SaveResponseBodyCmd(filename string, body []byte) tea.Cmd {
	return func() tea.Msg {
		err := os.WriteFile(filename, body, 0644)
		if err != nil {
			return uimsg.ErrMsg{Err: fmt.Errorf("error saving file: %w", err)}
		}
		return uimsg.StatusMsg{Message: "Saved to " + filename, IsError: false}
	}
}
