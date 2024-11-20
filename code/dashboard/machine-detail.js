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
    render_machine_load()
    let info = cluster_info(CURRENT_CLUSTER)
    if (info.violators) {
        document.getElementById("violators_link").href=`violators.html?cluster=${CURRENT_CLUSTER}&host=${CURRENT_HOST}`
    } else {
        document.getElementById("violators_link").replaceChildren()
    }
    if (info.deadweight) {
        document.getElementById("deadweight_link").href=`deadweight.html?cluster=${CURRENT_CLUSTER}&host=${CURRENT_HOST}`
    } else {
        document.getElementById("deadweight_link").replaceChildren()
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
