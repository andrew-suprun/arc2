* move the model.Entry to the view
* use slices.Delete to delete element from slice; no need to force sorting after
* use slices.BinarySearch/Insert to reduce need for sorting
* use slices.Index when possible
* sort using slices.SortFunc and cmp.Compare
* switch log to slog
* ??? make File an interface, maybe
* fix renaming folder
* implement delete event
* show copy file progress bar
* "resolve all" key shortcut
* share copy stats between archivers
* show stats summary
* filtering
* keyboard shortcuts for selecting sorting order
* keepFile on folders
* make logging optional triggered by '-log' command line flag
* handle 'keep all' event 
* add descriptions to ScanErrors
* ??? Separate Scroll into Scroll and Sized
* ??? store hashes as hex encoded strings
