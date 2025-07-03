package kinds

manifest: {
	appName: "settings"
  groupOverride: "settings.grafana.app"
	kinds: [setting]
	// extraPermissions: {
	// 	accessKinds: [
	// 		{
	// 			group: "settings.grafana.app"
	// 			resource: "settings"
	// 			actions: ["get","list","watch"]
	// 		}
	// 	]
	// }
}
