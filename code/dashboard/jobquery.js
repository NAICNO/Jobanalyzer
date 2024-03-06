// Logic for jobquery.html.

function setup() {
    makeTable()
    populateFromParameters()
    addProfilingHooks()
}

// The representation of "true" is a hack but it's determined by the server, so live with it.  Note
// the initial `=`.
const trueVal = "=xxxxxtruexxxxx"

// On startup - if there are parameters, populate the form elements from them
function populateFromParameters() {

    // The thing with "first" is a hack to work around that some browsers retain the form fields on
    // reload, even when there are URL parameters.
    function appendTo(elt, first, value) {
        if (elt.value != "" && !first) {
            elt.value += "," + value
        } else {
            elt.value = value
        }
    }

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

var theTable;

function makeTable() {
    function linkToJob(row, field) {
        let clusterVal = document.getElementById("clustername").value.trim()
        let fromVal = document.getElementById("fromdate").value.trim()
        let toVal = document.getElementById("todate").value.trim()
        let a = document.createElement("A")
        a.textContent = row[field.tag]
        a.href = makeProfileURL(clusterVal, fromVal, toVal, row)
        a.target = "_blank"
        return a
    }

    function breakText(row, field) {
        // Don't break at spaces that exist, but at commas.  Generally this has the effect of
        // keeping duration and timestamp fields together and breaking the command field apart.
        //
        // TODO: This is not ideal, b/c it breaks within node name ranges, we can fix that.
        return String(row[field.tag]).replaceAll(" ", "\xA0").replaceAll(",", ", ")
    }

    function percentage(row, field) {
        return (row[field.tag]/100).toFixed(1)
    }

    theTable = new Table(
        document.getElementById("joblist"),
        [{ name:    "Job#",
           tag:     "job",
           sort:    "numeric",
           display: linkToJob },

         { name:    "User",
           tag:     "user" },

         { name:    "Node",
           tag:     "host",
           display: breakText },

         { name:    "Duration",
           tag:     "duration",
           sort:    "duration"},

         { name:    "Start",
           tag:     "start" },

         { name:    "End",
           tag:     "end" },

         { name:    "Peak #cores",
           tag:     "cpu-peak",
           sort:    "numeric",
           display: percentage },

         { name:    "Peak resident GB",
           tag:     "res-peak",
           sort:    "numeric" },

         { name:    "Peak virtual GB",
           tag:     "mem-peak",
           sort:    "numeric" },

         { name:    "Peak GPU cards",
           tag:     "gpu-peak",
           sort:    "numeric",
           display: percentage },

         { name:    "Peak GPU RAM GB",
           tag:     "gpumem-peak",
           sort:    "numeric" },

         { name:    "Command",
           tag:     "cmd",
           display: breakText } ]
    )
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
        for ( let node of encodeStrings(splitMultiPattern(nodeVal))) {
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
    case "either": break;
    case "some": query += "&some-gpu" + trueVal; break;
    case "none": query += "&no-gpu" + trueVal; break;
    }
    let fmt = encodeURIComponent("json," + theTable.fields.map(x => x.tag).join(","))
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
	query += `&min-cpu-peak=${n*100}`
    }
    if (minPeakRamVal != "") {
	let n = parseInt(minPeakRamVal)
	if (!isFinite(n) || n < 0) {
            message("Invalid `min-peak-mem` value, must be finite and nonnegative")
	    return
	}
	query += `&min-res-peak=${n}`
    }

    // Display a URL that will take us back to this window with the form fields filled in
    message(query.replace("/jobs", window.location.origin + window.location.pathname))

    fetch(query).
	then(response => response.json()).
	then(function (data) {
            theTable.sortAndRepopulateFromData(
                data,
                "end",
                "descending")
        }).
        catch(function (e) {
            console.log(e)
            message("Query failed - is your cluster name correct?")
        })
}

var profilingHooks = false

function addProfilingHooks(name) {
    if (!profilingHooks) {
        // Add event handlers so that when somebody changes the profile settings, all the URLs in the
        // table are rewritten.  This is pretty gross in principle but works OK.
        for (let name of profnames) {
            document.getElementById("p" + name).addEventListener("change", recomputeURLs)
        }
        profilingHooks = true
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
    // FIXME
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
    if (from) {
	query += `&from=${from}`
    }
    if (to) {
	query += `&to=${to}`
    }
    return query
}

// Split s at "," and return non-empty trimmed strings

function splitAndEncode(s) {
    return encodeStrings(s.split(","))
}

function encodeStrings(vals) {
    let result = []
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
