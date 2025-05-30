// Code generated - EDITING IS FUTILE. DO NOT EDIT.

export interface Item {
	// type of the item.
	type: "dashboard_by_tag" | "dashboard_by_uid";
	// Value depends on type and describes the playlist item.
	//  - dashboard_by_tag: The value is a tag which is set on any number of dashboards. All
	//  dashboards behind the tag will be added to the playlist.
	//  - dashboard_by_uid: The value is the dashboard UID
	value: string;
}

export const defaultItem = (): Item => ({
	type: "dashboard_by_tag",
	value: "",
});

export interface Spec {
	title: string;
	interval: string;
	items: Item[];
}

export const defaultSpec = (): Spec => ({
	title: "",
	interval: "",
	items: [],
});

