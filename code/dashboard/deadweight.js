let CURRENT_HOST = ""

function deadweight_onload() {
    let params = new URLSearchParams(document.location.search)
    CURRENT_HOST = params.get("host")
    rewriteTitle(CURRENT_HOST)
    let info = cluster_info(CURRENT_CLUSTER)
    let hostFilter = null
    if (CURRENT_HOST != "") {
        hostFilter = function (d) { return d["hostname"] == CURRENT_HOST }
    }
    if (info.deadweight) {
	render(hostFilter)
    }
}

function render(filter) {
    rewriteTitle()
    let info = cluster_info(CURRENT_CLUSTER)
    if (!info.deadweight) {
        return
    }
    var fields = [{name: "Host", tag: "hostname"},
		  {name: "User", tag: "user"},
		  {name: "Job",  tag: "id"},
		  {name: "Command", tag:"cmd"},
		  {name: "First seen", tag:"started-on-or-before"},
		  {name: "Last seen", tag:"last-seen"}]
    render_table_from_report(tag_file("deadweight-report.json"),
			     fields,
			     document.getElementById("report"),
			     cmp_string_fields("last-seen", true),
                             filter)
}
