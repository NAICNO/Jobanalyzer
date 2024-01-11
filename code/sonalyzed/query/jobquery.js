// This code is very ad-hoc.  Some could be shared with the code in ../../dashboard.  Much could
// probably be extracted from a framework, if we were to use that.

// The representation of "true" is a hack but it's determined by the server, so live with it.  Note
// the initial `=`.
const trueVal = "=xxxxxtruexxxxx"

// On startup - if there are parameters, populate the form elements from them
function populateFromParameters() {
    let params = new URLSearchParams(document.location.search)
    let users = true
    let hosts = true
    let jobs = true
    for ( let [key, value] of params ) {
        // The keys are what they would also be for the query to the server
        switch (key) {
        case "cluster":
            document.getElementById("clustername").value = value
            break
        case "user":
            appendTo(document.getElementById("username"), users, value)
            users = false
            break
        case "host":
            appendTo(document.getElementById("nodename"), hosts, value)
            hosts = false
            break
        case "job":
            appendTo(document.getElementById("jobid"), jobs, value)
            jobs = false
            break
        case "from":
            document.getElementById("fromdate").value = value
            break
        case "to":
            document.getElementById("todate").value = value
            break
        case "min-runtime":
            document.getElementById("min-runtime").value = value
            break
        case "min-peak-cpu":
            document.getElementById("min-peak-cpu").value = value
            break
        case "min-peak-mem":    // [sic]
            document.getElementById("min-peak-ram").value = value // [sic]
            break
        case "some-gpu":
            document.getElementById("some-gpu").checked = true
            break
        case "no-gpu":
            document.getElementById("no-gpu").checked = true
            break
        }
    }
}

// The thing with "first" is a hack to work around that some browsers retain the form fields on
// reload, even when there are URL parameters.

function appendTo(elt, first, value) {
    if (elt.value != "" && !first) {
        elt.value += "," + value
    } else {
        elt.value = value
    }
}

var names = ["job","user","host","duration","start","end","cpu-peak","mem-peak","gpu-peak","gpumem-peak","cmd"]
var sorting = {
    "job": "numeric",
    "cpu-peak": "numeric",
    "mem-peak": "numeric",
    "gpu-peak": "numeric",
    "gpumem-peak": "numeric",
    "duration": "duration",
}

// In response to the "Select jobs" button - read and validate the form elements, post a query.
function selectJobs() {
    // Clear error message
    message("")

    // Validate fields and construct the query.
    let clusterVal = document.getElementById("clustername").value.trim()
    let userVal = document.getElementById("username").value.trim()
    let nodeVal = document.getElementById("nodename").value.trim()
    let jobVal = document.getElementById("jobid").value.trim()
    let fromVal = document.getElementById("fromdate").value.trim()
    let toVal = document.getElementById("todate").value.trim()
    let minRuntimeVal = document.getElementById("min-runtime").value.trim()
    let minPeakCpuVal = document.getElementById("min-peak-cpu").value.trim()
    let minPeakRamVal = document.getElementById("min-peak-ram").value.trim()
    let gpusel;
    if (document.getElementById("some-gpu").checked) {
	gpusel = "some"
    } else if (document.getElementById("no-gpu").checked) {
	gpusel = "none"
    } else {
	gpusel = "either"
    }
    if (clusterVal == "") {
	message("Cluster is required")
        return
    }
    let cluster = encodeURIComponent(clusterVal);
    let query = `/jobs?cluster=${cluster}`

    let users = splitAndEncode(userVal)
    if (users.length == 0) {
        users = ["-"]
    }
    for ( let user of users ) {
        query += `&user=${user}`
    }

    if (nodeVal != "") {
        for ( let node of splitAndEncode(nodeVal) ) {
	      query += `&host=${node}`
	}
    }

    if (jobVal != "") {
        for ( let job of splitAndEncode(jobVal) ) {
            let n = parseInt(job)
            if (!isFinite(n) || n <= 0) {
                message("Job ID must be finite and positive")
                return
            }
	    query += `&job=${job}`
	}
    }

    let validTime = /^(?:\d\d\d\d-\d\d-\d\d)|(?:\d+d)|(?:\d+w)$/
    let from, to
    if (fromVal != "") {
	if (validTime.exec(fromVal) == null) {
            message("Invalid `from` time, format is YYYY-MM-DD or Nw or Nd")
	    return
	}
	from = encodeURIComponent(fromVal)
	query += `&from=${from}`
    }
    if (toVal != "") {
	if (validTime.exec(toVal) == null) {
            message("Invalid `to` time, format is YYYY-MM-DD or Nw or Nd")
	    return
	}
	to = encodeURIComponent(toVal)
	query += `&to=${to}`
    }
    switch (gpusel) {
    case "none": break;
    case "some": query += "&some-gpu" + trueVal; break;
    case "none": query += "&no-gpu" + trueVal; break;
    }
    let fmt = encodeURIComponent("json," + names.join(","))
    query += `&fmt=${fmt}`
    if (minRuntimeVal != "") {
	if (/^(\d+w)?(\d+d)?(\d+h)?(\d+m)?$/.exec(minRuntimeVal) == null) {
            message("Invalid `min-runtime` value, format is WwDdHhMm with all parts optional but one required")
	    return
	}
	let v = encodeURIComponent(minRuntimeVal)
	query += `&min-runtime=${v}`
    }
    if (minPeakCpuVal != "") {
	let n = parseInt(minPeakCpuVal)
	if (!isFinite(n) || n < 0) {
            message("Invalid `min-peak-cpu` value, must be finite and nonnegative")
	    return
	}
	query += `&min-peak-cpu=${n}`
    }
    if (minPeakRamVal != "") {
	let n = parseInt(minPeakRamVal)
	if (!isFinite(n) || n < 0) {
            message("Invalid `min-peak-mem` value, must be finite and nonnegative")
	    return
	}
	query += `&min-peak-mem=${n}`
    }

    // Display a URL that will take us back to this window with the form fields filled in
    message(query.replace("/jobs", window.location.origin + window.location.pathname))

    // Fire off the query
    let tableState = {
        selected: "",
        direction: "",
    }
    fetch(query).
	then(response => response.json()).
	then(function (data) {
            sortDataBy(data, "end", "descending", tableState)
            repopulateTable(cluster, from, to, data, tableState)
        }).
        catch(function () {
            message("Query failed - is your cluster name correct?")
        })
}

var durationRe = /^(.*)d(.*)h(.*)m$/;

function sortDataBy(data, field, direction, tableState) {
    if (tableState.selected == field && direction == "opposite") {
        if (tableState.direction == "ascending") {
            direction = "descending"
        } else {
            direction = "ascending"
        }
    } else if (direction == "opposite") {
        direction = "ascending"
    }
    data.sort(function (a, b) {
        let x, y
        switch (sorting[field]) {
        case "numeric":
            x = parseFloat(a[field])
            y = parseFloat(b[field])
            break
        case "duration": {
            let m1 = durationRe.exec(a[field])
            let m2 = durationRe.exec(b[field])
            x = parseInt(m1[1])*24*60 + parseInt(m1[2])*60 + parseInt(m1[3])
            y = parseInt(m2[1])*24*60 + parseInt(m2[2])*60 + parseInt(m2[3])
            break
        }
        default:
            x = a[field]
            y = b[field]
            break
        }
        if (x < y) {
            return -1
        }
        if (x > y) {
            return 1
        }
        return 0
    })
    if (direction == "descending") {
        data.reverse()
    }
    tableState.selected = field
    tableState.direction = direction
}

function repopulateTable(cluster, from, to, data, tableState) {
    let tbl = document.getElementById("joblist")

    // Nuke any existing table
    tbl.replaceChildren()

    // Create the header
    let th = document.createElement("THEAD")
    let index = 0
    for (let name of names) {
	let td = document.createElement("TD")
        let marker = ""
        if (tableState.selected == name) {
            if (tableState.direction == "ascending")
                marker = " ^"
            else
                marker = " v"
        }
	td.textContent = name + marker
        td.addEventListener("click", function () {
            sortDataBy(data, name, "opposite", tableState)
            repopulateTable(cluster, from, to, data, tableState)
        })
	th.appendChild(td)
        index++
    }
    tbl.appendChild(th)

    // Create the rows
    for (let row of data) {
	let tr = document.createElement("TR")
	for (let name of names) {
	    let td = document.createElement("TD")
	    let text = row[name]
	    if (name.indexOf("/sec") != -1) {
		text = new Date(parseInt(text)*1000).toLocaleString()
	    }
            // Don't break at spaces that exist, but at commas.  Generally this has the effect of
            // keeping duration and timestamp fields together and breaking the command field apart.
            text = text.replaceAll(" ", "\xA0").replaceAll(",", ", ")
            if (name == "job") {
                let a = document.createElement("A")
	        a.textContent = text
	        a.href = makeProfileURL(cluster, from, to, row)
                a.target = "_blank"
                td.appendChild(a)
            } else {
	        td.textContent = text
            }
	    tr.appendChild(td)
	}
	tbl.appendChild(tr)
    }

    // Add event handlers so that when somebody changes the profile settings, all the URLs in the
    // table are rewritten.  This is pretty gross in principle but works OK.
    for (let name of profnames) {
        document.getElementById("p" + name).addEventListener("change", recomputeURLs)
    }
}

function recomputeURLs() {
    let profname = "cpu"
    for (let name of profnames) {
	if (document.getElementById("p" + name).checked) {
	    profname = name
	}
    }
    let as = document.getElementById("joblist").getElementsByTagName("a")
    for ( let a of as ) {
        // Replace "html,FACTOR" by "html,OTHERFACTOR" in the fmt= part of the URL
        // Use %2C for "," because the thing is already URL encoded
        a.href = a.href.replace(/html%2C[a-z]+/, "html%2C" + profname)
    }
}

// Pick up parameters and create a new window with the selected profile

var profnames = ["cpu", "mem", "gpu", "gpumem"]

function makeProfileURL(cluster, from, to, row) {
    // Oh, this is bad - for the ML cluster the host is necessary but for Fox et al it is wrong!!
    // Really the data could be self-identifying?
    let host = encodeURIComponent(row.host)
    let job = encodeURIComponent(row.job)
    let profname = "cpu"
    for (let name of profnames) {
	if (document.getElementById("p" + name).checked) {
	    profname = name
	}
    }
    let fmt = encodeURIComponent("html," + profname)
    let query = `/profile?cluster=${cluster}&job=${job}&host=${host}&fmt=${fmt}`
    if (from !== undefined) {
	query += `&from=${from}`
    }
    if (to !== undefined) {
	query += `&to=${to}`
    }
    return query
}

// Split s at "," and return non-empty trimmed strings

function splitAndEncode(s) {
    let result = []
    let vals = s.split(",")
    for (let v of vals) {
	v = v.trim()
	if (v != "") {
            result.push(encodeURIComponent(v))
	}
    }
    return result
}

// Put a message into the message div

function message(msg) {
    document.getElementById("message").textContent = msg
}
