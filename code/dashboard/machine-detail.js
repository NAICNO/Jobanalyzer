// dash-shared.js must have been loaded before this

let machine_load_chart = null
let CURRENT_HOST = ""

function machine_detail_onload() {
    let params = new URLSearchParams(document.location.search)
    CURRENT_HOST = params.get("host")
    document.title = CURRENT_HOST + " machine details"
    if (!cluster_info(CURRENT_CLUSTER).hasDowntime) {
        document.getElementById("downtime_cluster").remove()
    }
    document.getElementById("jobquery_link").href =
        `jobquery.html?cluster=${CURRENT_CLUSTER}&host=${CURRENT_HOST}`
    with_systems_and_frequencies(function (systems, frequencies) {
	frequencies = [{text: "Moment-to-moment (last 24h)", value: "minutely"}, ...frequencies]
	populateDropdown(document.getElementById("frequency"), frequencies)
	// TODO: Probably vet the host against the systems
	render()
    })
}

function render() {
    // We can argue about 30 but note that 30 is what the text says, and 30 is what the report
    // contains (currently).
    let thirty_days_ago = Date.now() - 30*24*60*60*1000
    render_machine_load()
    let info = cluster_info(CURRENT_CLUSTER)
    if (info.violators) {
        render_violators(thirty_days_ago)
    } else {
        document.getElementById("violators_display").replaceChildren()
    }
    if (info.deadweight) {
        render_deadweight(thirty_days_ago)
    } else {
        document.getElementById("deadweight_display").replaceChildren()
    }
}

function render_machine_load() {
    let frequency = document.getElementById("frequency").value
    let show_data = document.getElementById("show_data").checked
    let show_downtime = document.getElementById("show_downtime")?.checked
    let chart_node = document.getElementById("machine_load")
    let desc_node = document.getElementById("system_description")

    with_chart_data(CURRENT_HOST, frequency, function (json_data) {
	if (machine_load_chart != null)
	    machine_load_chart.destroy()
	machine_load_chart = plot_system(json_data, chart_node, desc_node, show_data, show_downtime, true)
    })
}

function render_violators(thirty_days_ago) {
    violators_by_time(thirty_days_ago)
    violators_by_user(thirty_days_ago)
}

function violators_by_time(thirty_days_ago) {
    document.getElementById("violator_report_by_time").replaceChildren()
    render_violators_by_time(
	"violator_report_by_time",
	function (d) {
	    return CURRENT_HOST == d["hostname"] &&
		(parse_date(d["last-seen"]) >= thirty_days_ago || globalThis["TESTDATA"])
	})
}

function violators_by_user(thirty_days_ago) {
    document.getElementById("violator_report_by_user").replaceChildren()
    render_violators_by_user(
	"violator_report_by_user",
	function (d) {
	    return CURRENT_HOST == d["hostname"] &&
		(parse_date(d["last-seen"]) >= thirty_days_ago || globalThis["TESTDATA"])
	},
	CURRENT_HOST)
}

/*
    let fields = [{name: "Host", tag: "hostname"},
		  {name: "User", tag: "user"},
		  {name: "Job",  tag: "id"},
		  {name: "Command", tag:"cmd"},
		  {name: "First seen", tag:"started-on-or-before"},
		  {name: "Last seen", tag:"last-seen"},
		  {name: "CPU peak", tag:"cpu-peak"},
		  {name: "RCPU avg", tag:"rcpu-avg"},
		  {name: "RCPU peak", tag:"rcpu-peak"},
		  {name: "RMem avg", tag:"rmem-avg"},
		  {name: "RMem peak", tag:"rmem-peak"}]
    let elt = document.getElementById("violator_report_by_time")
    elt.replaceChildren()
    render_table_from_file(tag_file("violator-report.json"),
			   fields,
			   elt,
			   cmp_string_fields("last-seen", true),
			   function (d) {
			       return CURRENT_HOST == d["hostname"] &&
				   (parse_date(d["last-seen"]) >= thirty_days_ago || globalThis["TESTDATA"])
			   })
}
*/

function render_deadweight(thirty_days_ago) {
      let fields = [{name: "Host", tag: "hostname"},
		    {name: "User", tag: "user"},
		    {name: "Job",  tag: "id"},
		    {name: "Command", tag:"cmd"},
		    {name: "First seen", tag:"started-on-or-before"},
		    {name: "Last seen", tag:"last-seen"}]
    let elt = document.getElementById("deadweight_report")
    elt.replaceChildren()
    render_table_from_file(tag_file("deadweight-report.json"),
			   fields,
			   elt,
			   cmp_string_fields("last-seen", true),
			   function (d) {
			       return CURRENT_HOST == d["hostname"] &&
				   (parse_date(d["last-seen"]) >= thirty_days_ago || globalThis["TESTDATA"])
			   })
}

