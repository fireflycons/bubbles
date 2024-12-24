# Bubbles

Some components for [Bubble Tea](https://github.com/charmbracelet/bubbletea)
applications.

## xtable

Extended Table

A port of [table](https://github.com/charmbracelet/bubbles/tree/master#table) with additional functionality.
* Store metadata on rows. Good for attaching the source data for the row making it easier to perform operations on the selected row.
* Sort and Find methods.
* Ability to add row numbers as column zero.
* Ability to create a table directly from a slice of arbitrary structs implementing the Metadata interface.
* Ability to delete rows:
    * At the cursor position
    * By row index
    * By hash value (Metadata interface)
    * By object - passing a value that implements the Metadata interface
* Method to find the vertical offset of the selected row from the top of the visible rows in the table.

## messagebox

A simple message box overlay.

