//******************************************************************
//
// SearchN4L - a simple command line search tool
//
// Prepare, e.g.:
// cd examples
// ../src/N4L-db -u chinese*n4l Mary.n4l doors.n4l cluedo.n4l brains.n4l
//
//******************************************************************

package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
)

/*import (
	"fmt"
	"os"
	"strings"
	"flag"

	SST "SSTorytime"
)*/

//******************************************************************

var (
	ARROWS  []string
	CHAPTER string
	SUBJECT string
	CONTEXT []string
	VERBOSE bool
	BROWSE  bool
	EXPLORE bool
	LIMIT   int
)

//******************************************************************

func main() {

	Init()

	load_arrows := true
	ctx := SST.Open(load_arrows)

	if SUBJECT == "" {
		fmt.Println("\nTo browse everything use: --browse everything..\n")
		Usage()
		os.Exit(1)
	}

	Search(ctx, ARROWS, CHAPTER, CONTEXT, SUBJECT, LIMIT)

	SST.Close(ctx)
}

//**************************************************************

func Usage() {

	fmt.Printf("usage: searchN4L [-v] [-arrows=] [-chapter string] subject [context]\n")
	flag.PrintDefaults()

	os.Exit(2)
}

//**************************************************************

func Init() []string {

	flag.Usage = Usage

	verbosePtr := flag.Bool("v", false, "verbose")
	chapterPtr := flag.String("chapter", "any", "a optional string to limit to a chapter/section")
	arrowsPtr := flag.String("arrows", "", "a list of forward/outward arrows to start with")
	limitPtr := flag.Int("limit", 20, "an approximate limit on the number of items returned, where applicable")
	browsePtr := flag.Bool("browse", false, "browse through all items")
	explorePtr := flag.Bool("explore", false, "explore items")

	flag.Parse()
	args := flag.Args()

	if *verbosePtr {
		VERBOSE = true
	}

	if *browsePtr {
		BROWSE = true
	}

	if *explorePtr {
		EXPLORE = true
	}

	if *arrowsPtr != "" {
		ARROWS = strings.Split(*arrowsPtr, ",")
	}

	LIMIT = *limitPtr

	if *chapterPtr != "" {
		CHAPTER = *chapterPtr
	}

	if len(args) > 0 {
		SUBJECT = args[0]

		for c := 1; c < len(args); c++ {
			CONTEXT = append(CONTEXT, args[c])
		}

		if len(ARROWS) == 0 && len(args) < 1 {
			Usage()
			os.Exit(1)
		}
	}

	if CONTEXT == nil {
		CONTEXT = append(CONTEXT, "")
	}

	if ARROWS == nil {
		ARROWS = append(ARROWS, "")
	}

	if len(ARROWS) == 0 {
		Usage()
		os.Exit(1)
	}

	SST.MemoryInit()

	return args
}

//******************************************************************

func Search(ctx SST.PoSST, arrows []string, chapter string, context []string, searchtext string, limit int) {

	fmt.Println()
	fmt.Println("** PROVISIONAL SEARCH TOOL *************************************\n")
	fmt.Println("   Searching in chapter", chapter)
	fmt.Println("   With context", context)
	fmt.Println("   Selected arrows", arrows)
	fmt.Println("   Node filter", searchtext)
	fmt.Println("\n")

	if BROWSE && searchtext == "everything" {
		searchtext = ""
	}

	EventSearch(ctx, chapter, context, searchtext, limit)

	if EXPLORE {
		BroadByName(ctx, chapter, context, searchtext, arrows, limit)
	}

	if BROWSE {
		Systematic(ctx, chapter, context, searchtext, arrows, limit)
	}

	chaps := SST.GetDBChaptersMatchingName(ctx, "")
	ctxts := SST.GetDBContextsMatchingName(ctx, "")

	TOC(chaps, ctxts)
}

//******************************************************************

func EventSearch(ctx SST.PoSST, chaptext string, context []string, searchtext string, limit int) {

	var count int

	nptrs := SST.GetDBNodePtrMatchingName(ctx, searchtext, chaptext)

	for nptr := range nptrs {
		count++
		if count > limit {
			return
		}
		fmt.Print("\n", nptr, ": ")
		SST.PrintNodeOrbit(ctx, nptrs[nptr], 100)
	}
}

//******************************************************************

func BroadByName(ctx SST.PoSST, chaptext string, context []string, searchtext string, arrnames []string, limit int) {

	const maxdepth = 5

	var start_set []SST.NodePtr

	search_items := strings.Split(searchtext, " ")

	for w := range search_items {
		start_set = append(start_set, SST.GetDBNodePtrMatchingName(ctx, search_items[w], chaptext)...)
	}

	for start := range start_set {

		if start+1 > limit {
			return
		}

		for sttype := SST.NEAR; sttype <= SST.EXPRESS; sttype++ {

			name := SST.GetDBNodeByNodePtr(ctx, start_set[start])

			allnodes := SST.GetFwdConeAsNodes(ctx, start_set[start], sttype, maxdepth)

			if len(allnodes) > 1 {
				fmt.Println()
				fmt.Println("    -------------------------------------------")
				fmt.Printf("     #%d via %s connection\n", start+1, SST.STTypeName(sttype))
				fmt.Printf("     (search %s => hit %s)\n", searchtext, name.S)
				fmt.Println("    -------------------------------------------")

				for l := range allnodes {
					fullnode := SST.GetDBNodeByNodePtr(ctx, allnodes[l])

					if !strings.Contains(fullnode.Chap, chaptext) {
						continue
					}

					//fmt.Println("     - SSType",SST.STTypeName(sttype)," cone item: ",fullnode.S,", found in",fullnode.Chap)
					SST.PrintNodeOrbit(ctx, allnodes[l], SST.SCREENWIDTH)
				}

				alt_paths, path_depth := SST.GetFwdPathsAsLinks(ctx, start_set[start], sttype, maxdepth)

				if alt_paths != nil {

					fmt.Println("\n  ", SST.STTypeName(sttype), "stories in the forward cone ----------------------------------")

					for p := 0; p < path_depth; p++ {
						SST.PrintLinkPath(ctx, alt_paths, p, "\n   found", "", context)
					}
				}
				fmt.Printf("     (END %d)\n", start+1)
			}
		}
	}
}

//******************************************************************

func Systematic(ctx SST.PoSST, chaptext string, context []string, searchtext string, arrnames []string, limit int) {

	chaptext = strings.TrimSpace(chaptext)
	searchtext = strings.TrimSpace(searchtext)

	var arrows []SST.ArrowPtr

	if arrnames[0] == "" {
		fmt.Println("\nTo browse, you need to specify some arrows with -arrows=")
		os.Exit(-1)
	} else {
		fmt.Println("\nSystematic browsing of nodes anchoring arrows...")
	}

	for a := range arrnames {
		arr := SST.GetDBArrowByName(ctx, arrnames[a])
		arrows = append(arrows, arr)
	}

	// Just print section 1
	qnodes := SST.GetDBNodeContextsMatchingArrow(ctx, searchtext, chaptext, context, arrows, 1)

	var prev string
	var header []string
	var count int

	for q := range qnodes {

		count++

		if count > limit {
			break
		}

		if qnodes[q].Context != prev {
			prev = qnodes[q].Context
			header = SST.ParseSQLArrayString(qnodes[q].Context)
			Header(header, qnodes[q].Chapter)
		}

		result := SST.GetDBNodeByNodePtr(ctx, qnodes[q].NPtr)
		SearchStoryPaths(ctx, result.S, result.NPtr, arrows, result.Chap, context)
	}

}

//**************************************************************

func SearchStoryPaths(ctx SST.PoSST, name string, start SST.NodePtr, arrows []SST.ArrowPtr, chap string, context []string) {

	const maxdepth = 8

	fmt.Println("....................................................................................")

	cone, _ := SST.GetFwdPathsAsLinks(ctx, start, 1, maxdepth)
	ShowCone(ctx, cone, 1, chap, context)

	cone, _ = SST.GetFwdPathsAsLinks(ctx, start, -1, maxdepth)
	ShowCone(ctx, cone, 1, chap, context)

	cone, _ = SST.GetFwdPathsAsLinks(ctx, start, 2, maxdepth)
	ShowCone(ctx, cone, 1, chap, context)

	cone, _ = SST.GetFwdPathsAsLinks(ctx, start, -2, maxdepth)
	ShowCone(ctx, cone, 1, chap, context)

	cone, _ = SST.GetFwdPathsAsLinks(ctx, start, 3, maxdepth)
	ShowCone(ctx, cone, 1, chap, context)

	cone, _ = SST.GetFwdPathsAsLinks(ctx, start, -3, maxdepth)
	ShowCone(ctx, cone, 1, chap, context)

	cone, _ = SST.GetFwdPathsAsLinks(ctx, start, 0, maxdepth)
	ShowCone(ctx, cone, 1, chap, context)
}

//**************************************************************

func ShowCone(ctx SST.PoSST, cone [][]SST.Link, sttype int, chap string, context []string) {

	if len(cone) < 1 {
		return
	}

	for s := 0; s < len(cone); s++ {

		SST.PrintLinkPath(ctx, cone, s, " - ", chap, context)
	}

}

//**************************************************************

func Header(h []string, chap string) {

	if len(h) == 0 {
		return
	}

	fmt.Println("\n\n============================================================")
	fmt.Println("   In chapter: \"", chap, "\"\n")

	for s := range h {
		fmt.Println("   ::", h[s], "::")
	}

	fmt.Println("\n============================================================")
}

//**************************************************************

func TOC(chap, cont []string) {

	if len(chap) == 0 && len(cont) == 0 {
		return
	}

	fmt.Println("\n\n============================================================")
	fmt.Println("\n   Chapters: \n")

	for s := range chap {
		fmt.Println("   - ", chap[s])
	}

	fmt.Println("\n   Contexts: \n")

	for s := range cont {
		SST.NewLine(s)
		fmt.Printf(" %-19.20s ", cont[s])
	}

	fmt.Println("\n============================================================")
}
