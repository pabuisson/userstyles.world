package main

import (
	"userstyles.world/database"
	"userstyles.world/handlers"
	"userstyles.world/images"
	"userstyles.world/search"
	"userstyles.world/updater"
	"userstyles.world/utils"
)

func main() {
	utils.InitalizeCrypto()
	utils.InitializeValidator()
	database.Initialize()
	search.Initialize()
	updater.Initialize()
	images.Initialize()
	handlers.Initialize()
}
