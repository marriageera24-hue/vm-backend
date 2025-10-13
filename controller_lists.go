package main

import (
	"net/http"
	"os"

	"github.com/labstack/echo"
)

type ListsResponse struct {
	Sites      ListItems `json:"sites"`
	AssetTypes ListItems `json:"asset_types"`
	Config     ListItems `json:"config"`
	States     ListItems `json:"states"`
}

func listAPIHandler(ctx echo.Context) error {
	var (
		resp ListsResponse
	)

	// resp.Sites = GetSites()
	// resp.AssetTypes = GetAssetTypes()
	resp.Config = getConfigVars()
	resp.States = getStates()

	return ctx.JSON(http.StatusOK, resp)
}

func getConfigVars() ListItems {
	var cv ListItems

	cv = append(cv, ListItem{
		Label: "vm_base_url",
		Value: os.Getenv("VM_BASE_URL"),
	})

	return cv
}
