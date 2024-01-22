// The global TESTDATA will be defined and true if the page loads testflag.js first and the flag is
// set there.  If it is defined and true we will redirect file queries to the test-data/ dir, which
// has static test data; and other code may decide to change its filtering in response to the
// setting.

function compute_filename(fn) {
    if (globalThis["TESTDATA"]) {
        return "test-data/" + fn;
    } else {
        return "output/" + fn;
    }
}

// We use CURRENT_CLUSTER to hold the tag of the cluster we're currently operating on.  All pages
// are relative to one specific cluster.

var CURRENT_CLUSTER = (function () {
    let params = new URLSearchParams(document.location.search)
    let cluster = params.get("cluster")
    return cluster ? cluster : "ml"
})();

// Cluster-info returns information about the cluster.  This function must be manually updated
// whenever we add a cluster.  That is fixable - it could be data in a config file.

function cluster_info(cluster) {
    switch (cluster) {
    default:
        /*FALLTHROUGH*/
    case "ml":
        return {
            cluster,
            subclusters: ["nvidia"],
            name:"ML nodes",
            description:"UiO Machine Learning nodes",
            prefix:"ml-",
            policy:"Significant CPU usage without any GPU usage",
        }
    case "fox":
        return {
            cluster,
            subclusters: ["cpu","gpu","int","login"],
            name:"Fox",
            description:"UiO 'Fox' supercomputer",
            prefix:"fox-",
            policy:"(To be determined)",
        }
    case "saga":
        return {
            cluster,
            subclusters: ["login"],
            name:"Saga",
            description:"Sigma2 'Saga' supercomputer",
            prefix:"saga-",
            policy:"(To be determined)",
        }
    }
}

// Update the window title and the main document title with the cluster name.

function rewriteTitle(extra) {
    let info = cluster_info(CURRENT_CLUSTER)
    let name = info.name + (extra ? " (" + extra + ")" : "")
    document.title = document.title.replace("CLUSTER", name)
    let title_elt = document.getElementById("main_title")
    if (title_elt) {
        title_elt.textContent = title_elt.textContent.replace("CLUSTER", name)
    }
}

// Add a prefix to the file name based on the cluster.

function tag_file(fn) {
    let info = cluster_info(CURRENT_CLUSTER)
    return info.prefix + fn
}

// Returns a promise that fetches and unwraps JSON data.

function fetch_data_from_file(f) {
    return fetch(compute_filename(f)).
        then((response) => response.json())
}

// Load tables of host names (in the cluster) and measurement frequencies and invoke f on those
// tables.

function with_systems_and_frequencies(f) {
    fetch_data_from_file(tag_file("hostnames.json")).
        then(function (json_data) {
            let systems = json_data.map(x => ({text: x, value: x}))
            let frequencies = [{text: "Daily, by hour", value: "daily"},
                               {text: "Weekly, by hour", value: "weekly"},
                               {text: "Monthly, by day", value: "monthly"},
                               {text: "Quarterly, by day", value: "quarterly"}]
            f(systems, frequencies)
        })
}

// Load the chart data and invoke f on the resulting table.

function with_chart_data(hostname, frequency, f) {
    fetch_data_from_file(`${hostname}-${frequency}.json`).
        then(f)
}

// json_data has these fields
//   date - string - the time the data was generated
//   hostname - string - FQDN (ideally) of the host
//   tag - string - usually "daily", "weekly", "monthly", "quarterly"
//   bucketing - string - "hourly" or "daily"
//   labels - array of length N of string labels
//   rcpu - array of length N of data values
//   rmem - same
//   rgpu - same, may be null/absent
//   rgpumem - same, may be null/absent
//   downhost - same, or null; values are 0 or 1
//   downgpu - same, or null; values are 0 or 1
//   system - system descriptor, see further down
//
// chart_node is a CANVAS
//
// desc_node is usually a DIV

function plot_system(json_data, chart_node, desc_node, show_data, show_downtime, show_hostname) {

    // Clamp GPU data to get rid of occasional garbage, it's probably OK to do this even
    // if it's not ideal.
    let labels = json_data.labels
    let rcpu_data = json_data.rcpu
    let rmem_data = json_data.rmem
    let rgpu_data = json_data.rgpu ? json_data.rgpu.map(d => Math.min(d, 100)) : null
    let rgpumem_data = json_data.rgpumem

    // Downtime data are flags indicating that the host or gpu was down during specific periods -
    // during the hour / day starting with at the start time of the bucket.  To represent that in
    // the current plot, we carry each nonzero value forward to the next slot too, to get a
    // horizontal line covering the entire bucket.  To make that pretty, we delete the remaining
    // zero slots.
    let downhost_data, downgpu_data
    if (json_data.downhost) {
        let dh = json_data.downhost.map(d => d*15)
        for ( let i=dh.length-1 ; i > 0 ; i-- ) {
            if (dh[i-1] > 0) {
                dh[i] = dh[i-1]
            }
        }
	for ( let i=0 ; i < dh.length ; i++ ) {
	    if (dh[i] == 0) {
		delete dh[i]
	    }
	}
        downhost_data = dh
    }
    if (json_data.downgpu) {
        let dg = json_data.downgpu.map(d => d*30)
        for ( let i=dg.length-1 ; i > 0 ; i-- ) {
            if (dg[i-1] > 0) {
                dg[i] = dg[i-1]
            }
        }
	for ( let i=0 ; i < dg.length ; i++ ) {
	    if (dg[i] == 0) {
		delete dg[i]
	    }
	}
        downgpu_data = dg
    }

    // Scale the chart.  Mostly this is now for the sake of rmem_data, whose values routinely
    // go over 100%.
    let maxval = Math.max(...rcpu_data, ...rmem_data)
    if (rgpu_data) {
	maxval = Math.max(maxval, ...rgpu_data)
    }
    if (rgpumem_data) {
	maxval = Math.max(maxval, ...rgpumem_data)
    }
    maxval = Math.max(maxval, 100)

    let datasets = []
    if (show_data) {
        datasets.push({ label: 'CPU%', data: rcpu_data, borderWidth: 2 },
                      { label: 'RAM%', data: rmem_data, borderWidth: 2 })
	if (rgpu_data) {
            datasets.push({ label: 'GPU%', data: rgpu_data, borderWidth: 2 })
	}
	if (rgpumem_data) {
            datasets.push({ label: 'VRAM%', data: rgpumem_data, borderWidth: 2 })
	}
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
	let desc = reformat_descriptions(json_data.system.description)
        desc_node.innerText = (show_hostname ? json_data.system.hostname + ": " : "") + desc
    }

    return my_chart;
}

// The description can contain multiple descriptions separated by "|||". We merge descriptions that
// are the same (as this is common).
function reformat_descriptions(description) {
    // Map from string description to numeric count
    let newdescs = {}
    for (let x of description.split("|||")) {
	if (x in newdescs) {
	    newdescs[x]++
	} else {
	    newdescs[x] = 1
	}
    }
    let xs = [];
    for (let x in newdescs) {
	xs.push([newdescs[x], x])
    }
    xs.sort(function (a, b) {
	if (a[0] == b[0]) {
	    if (a[1] < b[1]) {
		return -1
	    }
	    if (a[1] > b[1]) {
		return 1
	    }
	    return 0
	}
	return b[0] - a[0]
    })
    let desc = ""
    for (let [n, d] of xs) {
	if (desc != "") {
	    desc += "\n"
	}
	desc += n + "x " + d
    }
    return desc
}

// `dd` should be a SELECT element. `vals` should be an array of {text, value} objects with string values.
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

// Create a table from the fields and attach it to the parent.  Return the table and table body
// elements.
function make_table(fields, parent) {
    let tbl = document.createElement("table")
    let thead = document.createElement("thead")
    for ( let {name,help} of fields ) {
        let th = document.createElement("th")
        let sp = document.createElement("span")
        if (help) {
            sp.title = help
        }
        sp.textContent = name
        th.appendChild(sp)
        thead.appendChild(th)
    }
    tbl.appendChild(thead)
    let tbody = document.createElement("tbody")
    tbl.appendChild(tbody)
    parent.appendChild(tbl)
    return [tbl, tbody]
}

// Returns a promise that will fetch and render data in a table, which is made here.
function render_table_from_file(file, fields, parent, cmp, filter) {
    let [tbl, tbody] = make_table(fields, parent)
    return fetch_data_from_file(file).
        then(data => render_table_from_data(data, fields, tbody, cmp, filter))
}

// The `data` correspond to the `fields` somehow and are inserted as rows into `tbody`.  Returns
// array of new TR rows.  Each cell is a SPAN inside a TD.
function render_table_from_data(data, fields, tbody, cmp, filter) {
    if (cmp != undefined) {
        data.sort(cmp)
    }
    let trs = []
    for (let d of data) {
        if (filter != undefined && !filter(d)) {
            continue
        }
        let tr = document.createElement("tr")
        for (let {tag} of fields) {
            let td = document.createElement("td")
            if (tag in d) {
                let sp = document.createElement("span")
                sp.textContent = String(d[tag]).replaceAll(",",", ")
                td.appendChild(sp)
            }
            tr.appendChild(td)
        }
        tbody.appendChild(tr)
        trs.push([d, tr])
    }
    return trs
}

function cmp_string_fields(field, flip) {
    return function(r1, r2) {
        let a = r1[field]
        let b = r2[field]
        if (a < b) {
            return flip ? 1 : -1
        }
        if (a > b) {
            return flip ? -1 : 1
        }
        return 0
    }
}

// Time format used pervasively by Jobanalyzer: "yyyy-mm-dd hh:mm"
let date_matcher = /^(\d\d\d\d)-(\d\d)-(\d\d) (\d\d):(\d\d)$/;

// Given a date matching date_matcher, return a Number representing that as a UTC time with
// millisecond precision.
function parse_date(s) {
    let ms = date_matcher.exec(s)
    if (ms == null) {
        throw new Error("Not a date (syntax): <" + s + ">")
    }
    let year = parseInt(ms[1], 10)
    let month = parseInt(ms[2], 10)
    let day = parseInt(ms[3], 10)
    let hour = parseInt(ms[4], 10)
    let min = parseInt(ms[5], 10)
    if (isNaN(year) || isNaN(month) || isNaN(day) || isNaN(hour) || isNaN(min)) {
        throw new Error("Not a date (value): " + s)
    }
    return Date.UTC(year, month-1, day, hour, min)
}

// These work for strings, too

function min(a, b) {
    return a < b ? a : b
}

function max(a, b) {
    return a > b ? a : b
}

// Fixed-width field

function fix(n, v) {
    v = String(v)
    if (v.length >= n) {
	return v.substring(0,n)
    }
    return v + spaces(n-v.length)
}

var spc = " "

function spaces(n) {
    while (spc.length < n) {
	spc = spc + spc
    }
    return spc.substring(0, n)
}

