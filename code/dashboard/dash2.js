// dashboard.js must have been loaded before this one.

"use strict";

// Colors for annotations
let col_DOWN = "tomato"             // a bit better contrast than "red"
let col_WORKINGHARD = "deepskyblue" // 75%
let col_WORKING = "lightskyblue"    // 50%
let col_COASTING = "lightcyan"      // 25%

// Where to go when clicking on a row, this is opened with ?host=<hostname> from the JSON datum
let machine_page = `machine-detail.html`

let timeout_minutes = 5

// At-a-glance data are generated by naicreport, every 5min ideally.  In addition to the fields
// below there are three fields "recent", "longer", and "long" that give the values of those
// quantities in minutes, and an optional field "tag" that is a human-consumable string that
// briefly describes the data (eg, "ML Nodes" to state that the data are from the ML nodes,
// we assume different clusters have different descriptions).

let fields = [
    // FQDN
    {name: "Host", tag: "hostname"},

    // From the last record on the host.  0=ok, 1=error (maybe more codes later).
    {name: "CPU\nstatus", tag: "cpu_status", help:"0=up, 1=down"},
    {name: "GPU\nstatus", tag: "gpu_status", help:"0=up, 1=down"},

    // Unique users in the period.  This will never be greater than jobs; a user can have
    // several jobs, but not zero, and jobs can only have one user.
    {name: "Users\n(recent)", tag: "users_recent", help:"Unique users running jobs"},
    {name: "Users\n(longer)", tag: "users_longer", help:"Unique users running jobs"},

    // Unique jobs running within the period.
    {name: "Jobs\n(recent)", tag: "jobs_recent", help:"Jobs big enough to count"},
    {name: "Jobs\n(longer)", tag: "jobs_longer", help:"Jobs big enough to count"},

    // Relative to system information.
    {name: "CPU%\n(recent)", tag: "cpu_recent", help:"Running average"},
    {name: "CPU%\n(longer)", tag: "cpu_longer", help:"Running average"},
    {name: "Mem%\n(recent)", tag: "mem_recent", help:"Running average"},
    {name: "Mem%\n(longer)", tag: "mem_longer", help:"Running average"},
    {name: "GPU%\n(recent)", tag: "gpu_recent", help:"Running average"},
    {name: "GPU%\n(longer)", tag: "gpu_longer", help:"Running average"},
    {name: "GPUMEM%\n(recent)", tag: "gpumem_recent", help:"Running average"},
    {name: "GPUMEM%\n(longer)", tag: "gpumem_longer", help:"Running average"},

    // Number of *new* violators and zombies encountered in the period, as of the last
    // generated report.  This currently changes rarely.
    {name: "Violators\n(new)", tag: "violators_long", help: "New jobs violating policy"},
    {name: "Zombies\n(new)", tag: "zombies_long", help: "New defunct and zombie jobs"},
]

// Compute field offsets in field table
let offs = {}
let by_tag = {}
for (let i in fields) {
    offs[fields[i].tag] = i
    by_tag[fields[i].tag] = fields[i]
}

let reload_timer = null

function toggleReload() {
    let checked = document.getElementById("autorefresh").checked
    if (!checked && reload_timer != null) {
        clearTimeout(reload_timer)
        reload_timer = null
    } else if (checked && reload_timer == null) {
        reload_timer = setTimeout(function () {
            window.location.reload()
        }, timeout_minutes*60*1000)
    }
}

function setupLinks() {
    let info = cluster_info(CURRENT_CLUSTER)

    document.getElementById("violators_link").href=`violators.html?cluster=${CURRENT_CLUSTER}`
    document.getElementById("deadweight_link").href=`deadweight.html?cluster=${CURRENT_CLUSTER}`
    document.getElementById("jobquery_link").href=`${sonalyzedAddress()}/q/jobquery.html?cluster=${CURRENT_CLUSTER}`

    let subnames = info.subclusters
    let subs = document.getElementById("subclusters")
    if (subnames && subs) {
        subs.appendChild(document.createTextNode("Aggregates: "))
        for (let sn of subnames) {
            let a = document.createElement("A")
            let s = document.createElement("SPAN")
            s.textContent = sn
            a.appendChild(s)
            a.href = `subcluster.html?cluster=${CURRENT_CLUSTER}&subcluster=${sn}`
            subs.appendChild(a)
            let t = document.createTextNode("    ")
            subs.appendChild(t)
        }
    }
}

function setup() {
    rewriteTitle()
    toggleReload()
    setupLinks()
    render()
}

function render() {
    render_table_from_file(tag_file("at-a-glance.json"),
			   fields,
			   document.getElementById("report"),
			   sort_records).
        then(annotate_rows)
}

let working_fields = [
    "cpu_recent","cpu_longer",
    "mem_recent","mem_longer",
    "gpu_recent","gpu_longer",
    "gpumem_recent","gpumem_longer",
]

function annotate_rows(rows) {
    // Defaults, updated in the loop
    let recent_minutes = 30
    let longer_minutes = 12*60
    let long_minutes = 24*60

    // rows is an array of [json-datum, row-element] pairs
    // each row has exactly the fields above, offsets are computed above too
    // each cell is a SPAN inside a TD
    for ( let [d,r] of rows ) {
	let link = document.createElement("A")
	link.href = `${machine_page}?cluster=${CURRENT_CLUSTER}&host=${d["hostname"]}`
	link.appendChild(r.children[offs.hostname].firstChild) // The SPAN is moved to the A
	r.children[offs.hostname].appendChild(link)            // Insert A into TD
        if (d.cpu_status != 0) {
            r.style.backgroundColor = col_DOWN
            continue
        }
        if (d.gpu_status != 0 && d.gpu_status != undefined) {
            r.children[offs.gpu_status].style.backgroundColor = col_DOWN
        }
        for ( let n of working_fields ) {
            switch (true) {
            case d[n] >= 75:
                r.children[offs[n]].style.backgroundColor = col_WORKINGHARD
                break
            case d[n] >= 50:
                r.children[offs[n]].style.backgroundColor = col_WORKING
                break
            case d[n] >= 25:
                r.children[offs[n]].style.backgroundColor = col_COASTING
                break
            }
        }
        if ("machine" in d) {
            r.children[offs.hostname].children[0].title = d.hostname + ": " + d.machine
        }
        recent_minutes = d.recent
        longer_minutes = d.longer
        long_minutes = d.long
    }

    document.getElementById("recent_defn").textContent = sanetime(recent_minutes)
    document.getElementById("longer_defn").textContent = sanetime(longer_minutes)
    document.getElementById("long_defn").textContent = sanetime(long_minutes)
}

function sanetime(mins) {
    if (mins < 60) {
        return mins + " mins"
    }
    if (mins % 60 == 0) {
        return (mins / 60) + " hrs"
    }
    return Math.round((mins / 60) * 10) / 10 + " hrs"
}

function sort_records(r1, r2) {
    // failing cpus are sorted higher
    // failing gpus are sorted higher
    // then non-failing systems
    // within each group, by hostname
    if (r1["cpu_status"] != r2["cpu_status"]) {
        return r2["cpu_status"] - r1["cpu_status"]
    }
    if (r1["gpu_status"] != r2["gpu_status"]) {
        return r2["gpu_status"] - r1["gpu_status"]
    }
    if (r1["hostname"] < r2["hostname"]) {
        return -1
    }
    if (r1["hostname"] > r2["hostname"]) {
        return 1
    }
    return 0
}