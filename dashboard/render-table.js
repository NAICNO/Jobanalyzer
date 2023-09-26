function render_table(file, fields, parent, cmp) {
    let tbl = document.createElement("table")
    let thead = document.createElement("thead")
    for ( let {name} of fields ) {
	let th = document.createElement("th")
	th.innerText = name
	thead.appendChild(th)
    }
    tbl.appendChild(thead)
    let tbody = document.createElement("tbody")
    tbl.appendChild(tbody)
    parent.appendChild(tbl)
    fetch("output/" + file).then(response => response.json()).then(data => do_render_table(data, fields, tbody, cmp))
}

function do_render_table(data, fields, tbody, cmp) {
    if (cmp != undefined) {
	data.sort(cmp)
    }
    for (let d of data) {
	let tr = document.createElement("tr")
	for (let {tag} of fields) {
	    let td = document.createElement("td")
	    if (tag in d) {
		td.innerText = String(d[tag])
	    }
	    tr.appendChild(td)
	}
	tbody.appendChild(tr)
    }
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


    
