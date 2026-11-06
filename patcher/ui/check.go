package ui

import (
	"errors"
	"fmt"
	"slices"
	"strings"

	"anamod/internal/install"
	"fyne.io/fyne/v2/theme"
)

const mostMismatchesShown = 6

func (p *patcher) recheck() {
	if p.root == "" {
		return
	}
	var result install.Result
	var err error
	root, pack := p.root, p.pack
	p.background("Checking the game files...", func() {
		result, err = install.Check(root, pack.Manifest)
	}, func() {
		if p.pack != pack {
			p.recheck()
			return
		}
		p.showCheck(result, err)
	})
}

func (p *patcher) showCheck(result install.Result, err error) {
	var wrongRegion *install.WrongRegionError
	switch {
	case errors.As(err, &wrongRegion):
		fmt.Fprintf(p.log, "%v\n", err)
		p.result = nil
		p.showPick()
		p.setStatus(theme.ErrorIcon(), "Wrong region: "+wrongRegion.Detail(), "")
		return
	case err != nil:
		fmt.Fprintf(p.log, "Check failed: %v\n", err)
		p.result = nil
		p.showPick()
		p.setStatus(theme.ErrorIcon(), "Couldn't check this folder", sentence(err))
		return
	}
	p.result = &result
	p.rememberGame()
	p.showOptions(result.Game)
	p.showGame()
	fmt.Fprintf(p.log, "Check: %s, %s\n", result.Game.Name, result.State)
	for _, m := range result.Mismatches {
		found := m.Found
		if m.Missing {
			found = "missing"
		}
		fmt.Fprintf(p.log, "  %s: found %s, expected %s\n", m.Path, found, m.Expected)
	}
	switch result.State {
	case install.Vanilla:
		p.setStatus(theme.ConfirmIcon(), result.Game.Name, "Vanilla -- AnaMod Not Installed!")
	case install.Installed:
		record := result.Record
		for id, check := range p.checks {
			check.SetChecked(slices.Contains(record.Options, id))
		}
		detail := "AnaMod " + record.AnamodVersion + " installed"
		if p.packIsNewer(record) {
			detail += ". AnaMod " + p.pack.Manifest.AnamodVersion + " is ready to install"
		}
		p.setStatus(theme.ConfirmIcon(), result.Game.Name, detail)
	case install.ChangedSinceInstall:
		p.setStatus(theme.WarningIcon(), result.Game.Name,
			fmt.Sprintf("AnaMod %s is installed, but %s since then:\n%s",
				result.Record.AnamodVersion, countFiles(len(result.Mismatches), "has changed",
					"have changed"), listMismatches(result.Mismatches)))
	case install.SomethingElse:
		p.setStatus(theme.ErrorIcon(), "Unknown Game State",
			fmt.Sprintf("AnaMod needs a clean THP8 folder... another mod may be installed. %s:\n%s",
				countFiles(len(result.Mismatches), "doesn't match the vanilla game",
					"don't match the vanilla game"),
				listMismatches(result.Mismatches)))
	}
}

func countFiles(n int, one, many string) string {
	if n == 1 {
		return "1 file " + one
	}
	return fmt.Sprintf("%d files %s", n, many)
}

func listMismatches(mismatches []install.Mismatch) string {
	var lines []string
	for i, m := range mismatches {
		if i == mostMismatchesShown {
			lines = append(lines, fmt.Sprintf("and %d more (all of them are in the log)",
				len(mismatches)-i))
			break
		}
		found := "changed"
		if m.Missing {
			found = "missing"
		}
		lines = append(lines, fmt.Sprintf("%s (%s)", strings.TrimPrefix(m.Path, usrdirPrefix),
			found))
	}
	return strings.Join(lines, "\n")
}
