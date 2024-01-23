// dashboard.js must be loaded before this

var policyNames = {
    "ml-cpuhog": {
	trigger: "Job uses more than 10% of system's CPU at peak, runs for at least 10 minutes, and uses no GPU at all",
	problem: "ML nodes are for GPU jobs.  Job is in the way of other jobs that need GPU",
	remedy: "Move your work to a GPU-less system such as Fox or Light-HPC",
    }
}

function render() {
    rewriteTitle()

    let params = new URLSearchParams(document.location.search)
    // There must be a user parameter
    let user = params.get("user")
    if (!user) {
	console.log("Failed to find user")
    }
    // It's OK for there to be no host
    let host = params.get("host")
    let fields = [{name: "Host", tag: "hostname", width: 8},
		  {name: "Job",  tag: "id", width: 8},
		  {name: "Policy", tag: "policy"},
		  {name: "First seen", tag:"started-on-or-before", width: 16},
		  {name: "Last seen", tag:"last-seen", width: 16},
		  {name: "CPU% avg", tag:"rcpu-avg"},
		  {name: "CPU% peak", tag:"rcpu-peak"},
		  {name: "Virt% avg", tag:"rmem-avg"},
		  {name: "Command", tag:"cmd", width: -1},
		 ]
    fetch_data_from_file(tag_file("violator-report.json")).
	then(function (data) {
	    let violatedPolicies = {}
	    // Fixup old data and collect some metadata
	    for ( let r of data ) {
		for ( let f of fields ) {
		    if (f.tag == "policy" && r.policy === undefined) {
			r.policy = "ml-cpuhog"
		    }
		}
		violatedPolicies[r.policy] = true
	    }
	    let policies = ""
	    let gloss = host ? ` for host ${host}` : ""
	    for (let p in violatedPolicies) {
		policies += `  ${p}:
    Trigger: ${policyNames[p].trigger}
    Problem: ${policyNames[p].problem}
    Remedy:  ${policyNames[p].remedy}
`
	    }
	    let s = `
Hi,

This is a message from your friendly UiO systems administrator.

To ensure that computing resources are used in the best possible way,
we monitor how jobs are using the systems and ask users to move when
they are using a particular system in a way that is contrary to the
intended use of that system.

You are receiving this message because you have been running jobs
in just such a manner, as detailed below.  Please apply the suggested
remedies (usually this means moving your work to another system).


"${cluster_info(CURRENT_CLUSTER).name}" individual policy violator report${gloss}

Report generated on ${new Date()}

User:
  ${user}

Policies violated:
${policies}
(Times below are UTC, job numbers are derived from session leader if not running under Slurm)

`
	    data = data.filter((d) => d.user == user && (!host || d.hostname.indexOf(host) == 0))
	    data.sort(cmp_string_fields("last-seen", true))
	    let first = true
	    for (let f of fields) {
		if (!first) {
		    s += " "
		}
		let w = "width" in f ? f.width : 10;
		if (w != -1) {
		    s += fix(w, f.name)
		} else {
		    s += f.name
		}
		first = false
	    }
	    s += "\n\n"
	    for (let r of data) {
		first = true
		for (let f of fields) {
		    let v = r[f.tag]
		    if (!first) {
			s += " "
		    }
		    let w = "width" in f ? f.width : 10;
		    if (f.tag == "hostname") {
			v = v.split(".")[0]
		    }
		    if (w != -1) {
			s += fix(w, v)
		    } else {
			s += v
		    }
		    first = false
		}
		s += "\n"
	    }
	    document.getElementById("report").textContent = s
	})
}
