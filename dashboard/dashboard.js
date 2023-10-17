// json_data has these fields
//   date - string - the time the data was generated
//   hostname - string - FQDN (ideally) of the host
//   tag - string - usually "daily", "weekly", "monthly", "quarterly"
//   bucketing - string - "hourly" or "daily"
//   rcpu - array of per_point_data
//   rgpu - same
//   rmem - same
//   rgpumem - same
//   system - system descriptor, see further down
//
// per_point_data has two fields
//   x - usually a string - a label
//   y - the value
//
// chart_node is a CANVAS
//
// desc_node is usually a DIV

function plot_system(json_data, chart_node, desc_node, show_data, show_downtime) {

    // TODO: Some of the following data cleanup could move to the generating side, and indeed
    // this would seriously reduce data volume.  Bug 169.

    // Clamp GPU data to get rid of occasional garbage, it's probably OK to do this even
    // if it's not ideal.
    let labels = json_data.rcpu.map(d => d.x)
    let rcpu_data = json_data.rcpu.map(d => d.y)
    let rmem_data = json_data.rmem.map(d => d.y)
    let rgpu_data = json_data.rgpu.map(d => Math.min(d.y, 100))
    let downhost_data = json_data.downhost ? json_data.downhost.map(d => d.y*15) : null
    let downgpu_data = json_data.downgpu ? json_data.downgpu.map(d => d.y*30) : null
    let rgpumem_data = json_data.rgpumem.map(d => Math.min(d.y, 100))

    // Scale the chart.  Mostly this is now for the sake of rmem_data, whose values routinely
    // go over 100%.
    let maxval = Math.max(Math.max(...rcpu_data),
			  Math.max(...rmem_data),
			  Math.max(...rgpu_data),
			  Math.max(...rgpumem_data),
			  100)

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
