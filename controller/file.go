package controller

import v "arc/view"

func (f *file) setState(state v.State) {
	f.state = state
	f.folder.mergeState(state)
}
