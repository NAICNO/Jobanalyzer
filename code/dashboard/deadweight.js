function render() {
    rewriteTitle()
    var fields = [{name: "Host", tag: "hostname"},
		  {name: "User", tag: "user"},
		  {name: "Job",  tag: "id"},
		  {name: "Command", tag:"cmd"},
		  {name: "First seen", tag:"started-on-or-before"},
		  {name: "Last seen", tag:"last-seen"}]
    render_table_from_file(tag_file("deadweight-report.json"),
			   fields,
			   document.getElementById("report"),
			   cmp_string_fields("last-seen", true))
}
