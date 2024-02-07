// dash-shared.js must have been loaded before this

let machine_load_chart = null
let CURRENT_HOST = ""

// The chart name (ie CURRENT_HOST) needs to be the name of the subcluster I think, eg ml-nvidia,
// fox-cpu, fox-gpu, fox-int, this will turn into fox-int-weekly, from which fox-int-weekly.json.

function setup() {
    let params = new URLSearchParams(document.location.search)
    let cluster = params.get("cluster")
    let subcluster = params.get("subcluster")
    CURRENT_HOST = cluster + "-" + subcluster
    rewriteTitle(subcluster)
    render()
}

function render() {
    let chart_node = document.getElementById("machine_load")
    let desc_node = document.getElementById("system_description")
    with_chart_data(CURRENT_HOST, "weekly", function (json_data) {
	if (machine_load_chart != null)
	    machine_load_chart.destroy()
	machine_load_chart = plot_system(json_data, chart_node, desc_node, true, false, false)
    })
}
