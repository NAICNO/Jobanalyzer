// Subroutines to render policy violators in various ways.  Note this file is not specific to
// violators.html, it is also used by machine-detail.{html,js}.

// dashboard.js must be loaded before this

function render_violators_by_time(elt_id, filter) {
    var fields = [{name: "Host", tag: "hostname"},
		  {name: "User", tag: "user"},
		  {name: "Job",  tag: "id"},
		  {name: "Command", tag:"cmd"},
		  {name: "First seen", tag:"started-on-or-before"},
		  {name: "Last seen", tag:"last-seen"},
		  {name: "CPU peak", tag:"cpu-peak"},
		  {name: "CPU% avg", tag:"rcpu-avg"},
		  {name: "CPU% peak", tag:"rcpu-peak"},
		  {name: "Mem% avg", tag:"rmem-avg"},
		  {name: "Mem% peak", tag:"rmem-peak"},
		 ]
    render_table_from_file(
	tag_file("violator-report.json"),
	fields,
	document.getElementById(elt_id),
	cmp_string_fields("last-seen", true),
	filter
    )
}

function render_violators_by_user(elt_id, filter, filter_by_host) {
    var fields = [{name: "User", tag: "user"},
		  {name: "No. violations", tag: "count"},
		  {name: "First seen", tag:"earliest"},
		  {name: "Last seen", tag:"latest"},
		 ]
    let [tbl, tbody] = make_table(fields, document.getElementById(elt_id))
    fetch_data_from_file(tag_file("violator-report.json")).
	then(function (data) {
	    data = filter ? data.filter(filter) : data;
	    data = violators_user_view(data)
	    let rows =
		render_table_from_data(
		    data,
		    fields,
		    tbody,
		    cmp_string_fields("user", false),
		)
	    // For each row, make the user clickable with a handler that will pop open
	    // a window with the user's violations
	    let violator_page = "violator.html"
	    for ( let [d,r] of rows ) {
		// r is a TR containing a TD containing a SPAN containing the user name
		let username = r.children[0].firstChild.textContent;
		let link = document.createElement("A")
		let linktxt = `${violator_page}?user=${username}`
		if (filter_by_host) {
		    linktxt += `&host=${filter_by_host}`
		}
		link.href = linktxt
		// TODO: This knows the user is column 0
		link.appendChild(r.children[0].firstChild) // The SPAN is moved to the A
		r.children[0].appendChild(link)            // Insert A into TD
	    }
	})
}

function violators_user_view(violator_data) {
    // Map from user name to {user, count, earliest, latest, jobs} where earliest
    // is start of earliest job and latest is end of latest job.
    let users = {}
    for ( let r of violator_data ) {
	let u = users[r.user]
	if (u !== undefined) {
	    u.count++
	    u.earliest = min(u.earliest, r["started-on-or-before"])
	    u.latest = max(u.latest, r["last-seen"])
	    u.jobs.push(r)
	} else {
	    users[r.user] = {
		user: r.user,
		count: 1,
		earliest: r["started-on-or-before"],
		latest: r["last-seen"],
		jobs: [r],
	    }
	}
    }
    let result = []
    for ( let u in users ) {
	result.push(users[u])
    }
    return result
}
