function with_systems_and_frequencies(f) {
    fetch("output/hostnames.json").
	then((response) => response.json()).
	then(function (json_data) {
	    let systems = json_data.map(x => ({text: x, value: x}))
            let frequencies = [{text: "Daily, by hour", value: "daily"},
		               {text: "Weekly, by hour", value: "weekly"},
		               {text: "Monthly, by day", value: "monthly"},
		               {text: "Quarterly, by day", value: "quarterly"}]
            f(systems, frequencies)
        })
}

function with_chart_data(hostname, frequency, f) {
    fetch(`output/${hostname}-${frequency}.json`).then((response) => response.json()).then(f)
}

// json_data has these fields
//   date - string - the time the data was generated
//   hostname - string - FQDN (ideally) of the host
//   tag - string - usually "daily", "weekly", "monthly", "quarterly"
//   bucketing - string - "hourly" or "daily"
//   labels - array of length N of string labels
//   rcpu - array of length N of data values
//   rgpu - same
//   rmem - same
//   rgpumem - same
//   downhost - same, or null; values are 0 or 1
//   downgpu - same, or null; values are 0 or 1
//   system - system descriptor, see further down
//
// chart_node is a CANVAS
//
// desc_node is usually a DIV

function plot_system(json_data, chart_node, desc_node, show_data, show_downtime) {

    // Clamp GPU data to get rid of occasional garbage, it's probably OK to do this even
    // if it's not ideal.
    let labels = json_data.labels
    let rcpu_data = json_data.rcpu
    let rmem_data = json_data.rmem
    let rgpu_data = json_data.rgpu.map(d => Math.min(d, 100))
    let rgpumem_data = json_data.rgpumem

    // Downtime data are flags indicating that the host or gpu was down during specific periods -
    // during the hour / day starting with at the start time of the bucket.  To represent that in
    // the current plot, we carry each nonzero value forward to the next slot too, to get a
    // horizontal line covering the entire bucket.  This is far from pretty because we then get
    // slopes up to and down from the horizontal line from the preceding and following time slots.
    // That is bug #171.
    let downhost_data, downgpu_data
    if (json_data.downhost) {
        let dh = json_data.downhost.map(d => d*15)
        let dg = json_data.downgpu.map(d => d*30)
        for ( let i=dh.length-1 ; i > 0 ; i-- ) {
            if (dh[i-1] > 0) {
                dh[i] = dh[i-1]
            }
            if (dg[i-1] > 0 && dh[i] == 0) {
                dg[i] = dg[i-1]
            }
        }
        downhost_data = dh
        downgpu_data = dg
    }

    // Scale the chart.  Mostly this is now for the sake of rmem_data, whose values routinely
    // go over 100%.
    let maxval = Math.max(...rcpu_data, ...rmem_data, ...rgpu_data, ...rgpumem_data, 100)

    let datasets = []
    if (show_data) {
        datasets.push({ label: 'CPU%', data: rcpu_data, borderWidth: 2 },
	              { label: 'RAM%', data: rmem_data, borderWidth: 2 },
	              { label: 'GPU%', data: rgpu_data, borderWidth: 2 },
	              { label: 'VRAM%', data: rgpumem_data, borderWidth: 2 })
    }
    if (show_downtime) {
        if (downhost_data) {
            datasets.push( { label: "DOWN", data: downhost_data, borderWidth: 3 } )
        }
        if (downgpu_data) {
            datasets.push( { label: "GPU_DOWN", data: downgpu_data, borderWidth: 3 } )
        }
    }
    let my_chart = new Chart(chart_node, {
	type: 'line',
	data: {
	    labels,
	    datasets
	},
	options: {
	    scales: {
		x: {
		    beginAtZero: true,
		},
		y: {
		    beginAtZero: true,
		    // Add a little padding at the top, not sure what's a good amount
		    // but 10 is about the least we can do.
		    max: Math.floor((maxval + 10) / 10) * 10,
		}
	    }
	}
    })

    if (desc_node && "system" in json_data) {
	// This is a json object with these fields:
	//  - hostname: FQDN
	//  - description: human-readable text string
	//  - cpu_cores: num cores altogether (so, 2x14 hyperthreaded = 2x14x2 = 56)
	//  - mem_gb: gb of main memory
	//  - gpu_cards: num cards
	//  - gpumem_gb: total amount of gpu memory
	// Really the description says it all so probably enough to print that
	desc_node.innerText =
	    json_data.system.hostname + ": " + json_data.system.description
    }

    return my_chart;
}


// dd should be a SELECT element
// vals should be an array of {text, value} objects with string values

function populateDropdown(dd, vals) {
    for ( let { text, value } of vals ) {
	let opt = document.createElement("OPTION")
        // Firefox is happy with .label or .innerText here but Chrome insists on innerText.
        // Works with Safari too.
	opt.innerText = text
	opt.value = value
	dd.appendChild(opt)
    }
}
