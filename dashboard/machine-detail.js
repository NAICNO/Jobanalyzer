let machine_load_chart = null
let current_host = ""

function setup() {
    let params = new URLSearchParams(document.location.search)
    current_host = params.get("host")
    document.title = current_host + " machine details"
    with_systems_and_frequencies(function (systems, frequencies) {
	frequencies = [{text: "Moment-to-moment (last 6h)", value: "minutely"}, ...frequencies]
	populateDropdown(document.getElementById("frequency"), frequencies)
	// TODO: Probably vet the host against the systems
	render()
    })
}

function render() {
    render_machine_load()
    render_cpuhogs()
    render_deadweight()
}

function render_machine_load() {
    let frequency = document.getElementById("frequency").value
    let show_data = document.getElementById("show_data").checked
    let show_downtime = document.getElementById("show_downtime").checked
    let chart_node = document.getElementById("machine_load")
    let desc_node = document.getElementById("system_description")

    with_chart_data(current_host, frequency, function (json_data) {
	if (machine_load_chart != null)
	    machine_load_chart.destroy()
	machine_load_chart = plot_system(json_data, chart_node, desc_node, show_data, show_downtime)
    })
}

function render_cpuhogs() {
    let file = "cpuhog-report.json"
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
    let elt = document.getElementById("cpuhog_report")
    elt.replaceChildren()
    render_table(file,
                 fields,
                 elt,
                 cmp_string_fields("last-seen", true),
                 function (d) { return current_host == d["hostname"] })
}

function render_deadweight() {
      let file = "deadweight-report.json"
      let fields = [{name: "Host", tag: "hostname"},
		    {name: "User", tag: "user"},
		    {name: "Job",  tag: "id"},
		    {name: "Command", tag:"cmd"},
		    {name: "First seen", tag:"started-on-or-before"},
		    {name: "Last seen", tag:"last-seen"}]
    let elt = document.getElementById("deadweight_report")
    elt.replaceChildren()
    render_table(file,
                 fields,
                 elt,
                 cmp_string_fields("last-seen", true),
                 function (d) { return current_host == d["hostname"] })
}

