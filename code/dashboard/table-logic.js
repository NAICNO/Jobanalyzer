// General table logic with clickable column headers to sort by the column.
//
// This is good enough for jobquery but not quite good enough to replace the tables in the main
// dashboard.
//
// TODO (features):
// - major/minor sort: dash wants to sort failing systems before others.  this could be implemented
//   with an argument `major` to the constructor, an optional predicate that is applied to all rows
//   before the rows are resorted, and which must return true for those rows that should appear
//   before other rows, ie, it partitions the rows into two groups, which are sorted individually.
// - coloring for cells and rows that are special: again perhaps `major` could be used (major-sorted
//   rows are colored)?
//
// TODO (bling):
// - sorting arrow will not play nice with column headers containing breakable spaces, I guess we
//   could put it below the text?

"use strict";

// Create a new stateful table.
//
// `tableElt` is an HTML "TABLE" element that we'll manipulate the contents of.
//
// `fields` is an array of field descriptors, in the order to be displayed.  Each descriptor
// is an object with these members:
//
//  - "name": required, printable name for the column header
//  - "tag":  required, field name in row objects
//  - "help": optional, tooltip
//  - "sort": optional, a string describing how to compare the field.  Known types are "numeric"
//    (decimal numbers) and "duration" (matches durationRe below).  The default is "text".
//  - "display": optional, a function that takes the row and field object and returns a value that
//    is its displayable representation.  This must be either an HTMLElement (which is appended as a
//    child to the table cell) or a value (which is converted to string and becomes the textcontent
//    of a table cell)
//
//  The field objects can also have members that are private to the display function.

function Table(tableElt, fields) {
    if (!(tableElt instanceof HTMLTableElement)) {
        throw new Error("Not a TABLE element: " + tableElt)
    }

    // Public read-only
    this.tableElt = tableElt
    this.fields = fields
    this.sortField = ""
    this.sortDirection = ""

    // Private
    this.data = []
    this.sortOrder = {}
    for ( let f of fields ) {
        if (f.hasOwnProperty("sort") && (f.sort == "numeric" || f.sort == "duration")) {
            this.sortOrder[f.tag] = f.sort
        }
    }
    this.durationRe = /^(.*)d(.*)h(.*)m$/;
}

// Change `this.data`, sort it, and display it.
//
// `jsonData` is an array of objects with fields tagged as by `this.fields`.
//
// `sortField` is a field tag to sort by.
//
// `sortDirection` is the direction to sort: "ascending", "descending", "opposite".

Table.prototype.sortAndRepopulateFromData = function(jsonData, sortField, sortDirection) {
    this.data = jsonData
    this.sortAndRepopulate(sortField, sortDirection)
}

// Sort `this.data` and display it.
//
// `sortField` is a field tag to sort by.
//
// `sortDirection` is the direction to sort: "ascending", "descending", "opposite".

Table.prototype.sortAndRepopulate = function(sortField, sortDirection) {
    this.sortDataBy(sortField, sortDirection)
    this.repopulateTable()
}

// Reorder the rows of `this.data`.
//
// `sortField` is a field tag to sort by.
//
// `sortDirection` is the direction to sort: "ascending", "descending", "opposite".

Table.prototype.sortDataBy = function(sortField, sortDirection) {
    if (this.sortField == sortField && sortDirection == "opposite") {
        if (this.sortDirection == "ascending") {
            sortDirection = "descending"
        } else {
            sortDirection = "ascending"
        }
    } else if (sortDirection == "opposite") {
        sortDirection = "ascending"
    }
    let theTable = this
    this.data.sort(function (a, b) {
        let x, y
        switch (theTable.sortOrder[sortField]) {
        case "numeric":
            x = parseFloat(a[sortField])
            y = parseFloat(b[sortField])
            break
        case "duration": {
            let m1 = theTable.durationRe.exec(a[sortField])
            let m2 = theTable.durationRe.exec(b[sortField])
            x = parseInt(m1[1])*24*60 + parseInt(m1[2])*60 + parseInt(m1[3])
            y = parseInt(m2[1])*24*60 + parseInt(m2[2])*60 + parseInt(m2[3])
            break
        }
        default:
            x = a[sortField]
            y = b[sortField]
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
    if (sortDirection == "descending") {
        this.data.reverse()
    }
    this.sortField = sortField
    this.sortDirection = sortDirection
}

Table.prototype.repopulateTable = function() {
    // Nuke any existing table
    this.tableElt.replaceChildren()

    // Create the header
    let thead = document.createElement("THEAD")
    let index = 0
    for (let field of this.fields) {
        let th = document.createElement("TH")
        let marker = ""
        // TODO: This will look ugly for multi-line column headers and
        // may require some additional layout / styling.
        if (this.sortField == field.tag) {
            marker = this.sortDirection == "ascending" ? " ^" : " v"
        }
        if (field.hasOwnProperty("help")) {
            th.title = field.help
        }
        th.textContent = field.name + marker
        th.addEventListener("click", () => this.sortAndRepopulate(field.tag, "opposite"))
        thead.appendChild(th)
        index++
    }
    this.tableElt.appendChild(thead)

    // Create the rows
    for (let row of this.data) {
        let tr = document.createElement("TR")
        for (let field of this.fields) {
            let td = document.createElement("TD")
            let content = String(row[field.tag])
            if (field.hasOwnProperty("display")) {
                content = field.display(row, field)
            }
            if (content instanceof HTMLElement) {
                td.appendChild(content)
            } else {
                td.textContent = String(content)
            }
            tr.appendChild(td)
        }
        this.tableElt.appendChild(tr)
    }
}
