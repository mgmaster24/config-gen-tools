package ui

import (
	"fmt"
	"io"
)

// banner is nvimforge's own CLI startup art — distinct from the ASCII
// header baked into the generated Neovim config's snacks.nvim dashboard
// (internal/genconfig/templates/lua/plugins/snacks.lua.tmpl). The two are
// separate concerns: this one is only ever seen while running the
// nvimforge binary itself.
const banner = `
███╗   ██╗██╗   ██╗██╗███╗   ███╗███████╗ ██████╗ ██████╗  ██████╗ ███████╗
████╗  ██║██║   ██║██║████╗ ████║██╔════╝██╔═══██╗██╔══██╗██╔════╝ ██╔════╝
██╔██╗ ██║██║   ██║██║██╔████╔██║█████╗  ██║   ██║██████╔╝██║  ███╗█████╗
██║╚██╗██║╚██╗ ██╔╝██║██║╚██╔╝██║██╔══╝  ██║   ██║██╔══██╗██║   ██║██╔══╝
██║ ╚████║ ╚████╔╝ ██║██║ ╚═╝ ██║██║     ╚██████╔╝██║  ██║╚██████╔╝███████╗
╚═╝  ╚═══╝  ╚═══╝  ╚═╝╚═╝     ╚═╝╚═╝      ╚═════╝ ╚═╝  ╚═╝ ╚═════╝ ╚══════╝`

// PrintBanner writes nvimforge's startup banner to w, unless show is
// false (the config.Config.ShowBanner / --no-banner setting).
func PrintBanner(w io.Writer, show bool) {
	if !show {
		return
	}
	fmt.Fprintln(w, bannerStyle.Render(banner))
	fmt.Fprintln(w)
}
