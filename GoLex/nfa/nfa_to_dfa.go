package nfa

import (
	"fmt"
	"os"
)

const (
	DFA_MAX    = 254      //DFA 最多节点数
	F          = -1       //用于初始化跳转表
	MAX_CHARS  = 128      //128个ascii字符
	NCOLS      = 10       //每打印10个数字就换行
	SCLASS     = "static" //对应c语言中的static类型
	TYPE       = "unsigned char"
	NCLOS      = 16
	DTRAN_NAME = "Yy_nxt"
)

type ACCEPT struct {
	acceptString string //接收节点对应的执行代码字符串
	anchor       Anchor
}

type DFA struct {
	group        int  //后面执行最小化算法时有用
	mark         bool //当前节点是否已经设置好接收字符对应的边
	anchor       Anchor
	set          []*NFA //dfa节点对应的nfa节点集合
	state        int    //dfa 节点号码
	acceptString string
	isAccepted   bool
}

type NfaDfaConverter struct {
	nstates    int     //当前dfa 节点计数
	lastMarked int     //下一个需要处理的dfa节点
	dtrans     [][]int //dfa状态机的跳转表
	accepts    []*ACCEPT
	dstates    []DFA       //所有dfa节点的集合
	groups     [][]int     //用于dfa节点分区
	inGroups   []int       //根据节点值给出其所在分区
	numGroups  int         //当前分区数
	fp         *os.File    //创建lex.yy.cc文件
	colMap     map[int]int //映射跳转表中冗余的列
	rowMap     map[int]int //映射跳转表中冗余的行
}

func NewNfaDfaConverter() *NfaDfaConverter {
	n := &NfaDfaConverter{
		nstates:    0,
		lastMarked: 0,
		dtrans:     make([][]int, DFA_MAX),
		dstates:    make([]DFA, DFA_MAX),
		groups:     make([][]int, DFA_MAX),
		inGroups:   make([]int, DFA_MAX),
		colMap:     make(map[int]int, MAX_CHARS),
		rowMap:     make(map[int]int, DFA_MAX),
		numGroups:  0,
	}

	for i := range n.dtrans {
		n.dtrans[i] = make([]int, MAX_CHARS)
	}

	for i := range n.groups {
		n.groups[i] = make([]int, 0)
	}

	fp, err := os.Create("lex.yy.c")
	if err != nil {
		panic("lex.yy.c create error")
	}
	n.fp = fp
	return n
}

func (n *NfaDfaConverter) colEquiv(i int, j int) bool {
	//判断给定两列是否相同
	for row := 0; row < n.nstates; row++ {
		if n.dtrans[row][i] != n.dtrans[row][j] {
			return false
		}
	}

	return true
}

func (n *NfaDfaConverter) rowEquiv(compressed [][]int, i int, j int) bool {
	//判断给定两行是否相同
	iRow := compressed[i]
	jRow := compressed[j]
	for k := 0; k < len(iRow); k++ {
		if iRow[k] != jRow[k] {
			return false
		}
	}

	return true
}

func (n *NfaDfaConverter) colCopy(dest [][]int, srcCol int, destCol int) {
	//将原跳转表srcCol对应的列拷贝到dest数组中destCol对应的列
	rows := n.nstates
	for i := 0; i < rows; i++ {
		dest[i][destCol] = n.dtrans[i][srcCol]
	}
}

func (n *NfaDfaConverter) rowCopy(src [][]int, srcRow int, destRow int) {
	//将压缩表对应的行拷贝到跳转表
	n.dtrans[destRow] = src[srcRow]
}

func (n *NfaDfaConverter) reduce() {
	//压缩跳转表中冗余的列,save用于存储没有与前面列重复的列的下标
	save := make([]int, 0)
	current := -1  //指向没有与前面列重复的列
	colReduce := 0 //当前不重复的列要存储到该变量指定的位置
	colNum := len(n.dtrans[0])
	for true {
		i := current + 1
		for ; i < colNum; i++ {
			_, ok := n.colMap[i]
			if !ok {
				//当前列不还没有被标志为重复
				break
			}
		}

		if i >= colNum {
			//已经没有不重复的列了
			break
		}

		save = append(save, i)
		current = i
		n.colMap[i] = colReduce
		for j := i + 1; j < colNum; j++ {
			//如果当前j指向的列与current指向的列重复，那么则压缩它
			_, ok := n.colMap[j]
			if !ok && n.colEquiv(current, j) {
				//当前列对应到压缩后跳转表对应的colReduce列
				n.colMap[j] = colReduce
			}
		}
		colReduce += 1
	}

	compressed := make([][]int, n.nstates)
	for k := 0; k < n.nstates; k++ {
		compressed[k] = make([]int, colReduce)
	}
	//把save中存储的那些不重复的列拷贝到compressed数组
	for k := 0; k < colReduce; k++ {
		srcCol := save[k]
		destCol := k
		n.colCopy(compressed, srcCol, destCol)
	}

	//使用跟列压缩相同的方法来压缩行
	save = make([]int, 0)
	current = -1   //指向没有与前面列重复的行
	rowReduce := 0 //当前不重复的行要存储到该变量指定的位置
	for true {
		i := current + 1
		for ; i < n.nstates; i++ {
			_, ok := n.rowMap[i]
			if !ok {
				break
			}
		}

		if i >= n.nstates {
			//已经没有不重复的行
			break
		}

		save = append(save, i)
		current = i
		n.rowMap[i] = rowReduce
		for j := i + 1; j < n.nstates; j++ {
			_, ok := n.rowMap[j]
			if !ok && n.rowEquiv(compressed, current, j) {
				n.rowMap[j] = rowReduce
			}
		}
		rowReduce += 1
	}

	n.dtrans = make([][]int, rowReduce)
	for k := 0; k < len(save); k++ {
		srcRow := save[k]
		n.rowCopy(compressed, srcRow, k)
	}
}

func (n *NfaDfaConverter) reduceTesting() {
	n.dtrans = [][]int{
		{1, 1, 1, 2, 2, 3, 3},
		{1, 1, 1, 2, 2, 3, 3},
		{2, 2, 2, 3, 3, 4, 4},
		{2, 2, 2, 3, 3, 5, 5},
	}
	n.nstates = 4
	/*
			经过列压缩后应该为
			[1, 2, 3
			 1, 2, 3,
			 2, 3, 4,
			 2, 3, 5,]
			再经过行压缩后应该为
		    [1,2,3
		     2,3,4
		     2,3,5]
	*/
}

func (n *NfaDfaConverter) Squash() {
	//执行冗余行和列的压缩
	//用于测试,若不然要讲其注释掉
	//n.reduceTesting()

	n.reduce()
	//打印压缩后的跳转表用于测试
	//n.PrintCompressedDTran()
	n.printColMap()
	n.printRowMap()
}

func (n *NfaDfaConverter) DoSquash() {
	n.Squash()
	fmt.Fprintf(n.fp, "%s %s %s[%d][%d]=\n", SCLASS, TYPE, DTRAN_NAME, len(n.dtrans), len(n.dtrans[0]))
	n.printArray(n.dtrans)
	n.cnext("Yy_nxt")
}

func (n *NfaDfaConverter) printArray(array [][]int) {
	nrows := len(array)
	ncols := len(array[0])
	fmt.Fprintf(n.fp, "{\n")
	col := 0
	for i := 0; i < nrows; i++ {
		fmt.Fprintf(n.fp, "/*  %d */ {", i)
		for col = 0; col < ncols; col++ {
			fmt.Fprintf(n.fp, "%d", array[i][col])
			if col < ncols-1 {
				fmt.Fprintf(n.fp, ", ")
			}
			if col%NCOLS == NCOLS-1 && col != ncols-1 {
				fmt.Fprintf(n.fp, "\n        ")
			}
		}

		if col > NCLOS {
			fmt.Fprintf(n.fp, "\n        ")
		}

		fmt.Fprintf(n.fp, "}")
		if i < nrows-1 {
			fmt.Fprintf(n.fp, ",\n")
		} else {
			fmt.Fprintf(n.fp, " \n")
		}
	}

	fmt.Fprintf(n.fp, "};\n")
}

func (n *NfaDfaConverter) printComment(commnets []string) {
	//打印c语言中的注释
	fmt.Fprintf(n.fp, "\n/*--------------------------------------\n")
	for _, str := range commnets {
		fmt.Fprintf(n.fp, "* %s\n", str)
	}
	fmt.Fprintf(n.fp, "*/\n\n")
}

func (n *NfaDfaConverter) printMap(compressMap map[int]int) {
	//需要将map转换为int类型数组
	count := len(compressMap)
	mapArray := make([]int, count)
	for key, value := range compressMap {
		mapArray[key] = value
	}

	length := len(mapArray)
	for j := 0; j < length-1; j++ {
		fmt.Fprintf(n.fp, "%d,", mapArray[j])
		if j%NCOLS == NCOLS-1 {
			fmt.Fprintf(n.fp, "\n    ")
		}
	}
	fmt.Fprintf(n.fp, "%d\n};\n\n", mapArray[length-1])
}

func (n *NfaDfaConverter) printColMap() {
	//打印压缩后的列映射表
	text := []string{
		"The Yy_cmap[] and Yy_rmap arrays are used as follows:",
		"",
		" next_state= Yydtran[ Yy_rmap[current_state] ][ Yy_cmap[input_char] ];",
		"",
		"Character positions in the Yy_cmap array are:",
		"",
		"   ^@  ^A  ^B  ^C  ^D  ^E  ^F  ^G  ^H  ^I  ^J  ^K  ^L  ^M  ^N  ^O",
		"   ^P  ^Q  ^R  ^S  ^T  ^U  ^V  ^W  ^X  ^Y  ^Z  ^[  ^\\  ^]  ^^  ^_",
		"        !   \"   #   $   %   &   '   (   )   *   +   ,   -   .   /",
		"    0   1   2   3   4   5   6   7   8   9   :   ;   <   =   >   ?",
		"    @   A   B   C   D   E   F   G   H   I   J   K   L   M   N   O",
		"    P   Q   R   S   T   U   V   W   X   Y   Z   [   \\   ]   ^   _",
		"    `   a   b   c   d   e   f   g   h   i   j   k   l   m   n   o",
		"    p   q   r   s   t   u   v   w   x   y   z   {   |   }   ~   DEL",
	}
	n.printComment(text)
	fmt.Fprintf(n.fp, "%s %s Yy_cmap[%d]=\n{\n  ", SCLASS, TYPE, MAX_CHARS)
	n.printMap(n.colMap)
}

func (n *NfaDfaConverter) printRowMap() {
	fmt.Fprintf(n.fp, "%s  %s  Yy_rmap[%d]\n{\n    ", SCLASS, TYPE, len(n.dtrans))
	n.printMap(n.rowMap)
}

func (n *NfaDfaConverter) cnext(name string) {
	text := []string{
		"yy_next(state,c) is given the current state number and input",
		"character and evaluates to the next state.",
	}
	n.printComment(text)
	fmt.Fprintf(n.fp, "define yy_next(state, c) (%s[Yy_rmap[state]][Yy_cmap[c]])\n", name)
}

func (n *NfaDfaConverter) PrintCompressedDTran() {
	fmt.Printf("\n-------Compressed DTran Table------------\n")
	for i := 0; i < n.nstates; i++ {
		for j := 0; j < MAX_CHARS; j++ {
			mapRow := n.rowMap[i]
			mapCol := n.colMap[j]
			if n.dtrans[mapRow][mapCol] != F {
				fmt.Printf("from state %d jump to state %d with input: %s\n", i, n.dtrans[mapRow][mapCol], string(j))
			}
		}
	}
}

func (n *NfaDfaConverter) pairs(name string, threshold int, numbers int) int {
	/*
		生成压缩后的跳转表，并将其作为C代码进行输出。其中name对应压缩后数组在输出成代码时对应的
		数组变量名，threshold用于决定给定行是否需要压缩，如果一行中有效跳转的数量超过了threshold，
		那么就不用压缩，numbers决定数值对中，第一个元素使用字符表示还是ascii数值表示，如果members = 0,
		那么对应输入字符'0'，跳转对则为'0', 3, 如果members=1，那么跳转对为48,3，因为'0'对应ascii的值为48
	*/
	nrows := n.nstates
	numCells := 0 //数值对计数
	//开始输出第一部分
	for i := 0; i < nrows; i++ {
		ncols := len(n.dtrans[i])
		ntransation := 0 //统计有效跳转数
		for j := 0; j < ncols; j++ {
			if n.dtrans[i][j] != F {
				ntransation += 1
			}
		}

		if ntransation > 0 {
			//如果存在有效跳转则准备压缩,先输入压缩后数据对应的数组变量名
			fmt.Fprintf(n.fp, "%s %s %s%d[] = { ", SCLASS, TYPE, name, i)
			numCells += 1
			if ntransation > threshold {
				//如果有效跳转数超过了阈值就不要压缩，要不然数据没有减少反而会增加
				fmt.Fprintf(n.fp, "0, \n        ")
			} else {
				//先输入跳转数值对的数量
				fmt.Fprintf(n.fp, "%d, ", ntransation)
				if threshold > 5 {
					//如果阈值大于5，那么我们先输入换行，这么做是希望打印出来的内容看起来比较工整
					fmt.Fprintf(n.fp, "\n        ")
				}
			}

			nprinted := NCOLS
			ncommas := ntransation
			for j := 0; j < ncols; j++ {
				if ntransation > threshold {
					//有效跳转数超过了阈值，那么直接输出原来的行，因为此时压缩有可能会让数据量增大
					numCells += 1
					nprinted -= 1
					fmt.Fprintf(n.fp, "%d", n.dtrans[i][j])
					if j < ncols-1 {
						fmt.Fprintf(n.fp, ", ")
					}
				} else {
					//使用pair压缩算法来压缩行
					if n.dtrans[i][j] != F {
						if numbers > 0 {
							fmt.Fprintf(n.fp, "%d,%d", j, n.dtrans[i][j])
						} else {
							fmt.Fprintf(n.fp, "'%s',%d", string(j), n.dtrans[i][j])
						}

						nprinted -= 2
						ncommas -= 1
						if ncommas > 0 {
							fmt.Fprintf(n.fp, ", ")
						}
					}
				}

				if nprinted <= 0 {
					//输出换行
					fmt.Fprintf(n.fp, "\n          ")
					nprinted = NCOLS
				}
			}

			fmt.Fprintf(n.fp, "};\n")
		}
	}
	//结束输出第一部分

	//开始输出第二部分
	fmt.Fprintf(n.fp, "\n%s %s *%s[ %d ] = \n{\n    ", SCLASS, TYPE, name, nrows)
	nprinted := 10
	i := 0
	for i = 0; i < nrows-1; i++ {
		ntransations := 0
		ncols := len(n.dtrans[i])
		for j := 0; j < ncols; j++ {
			if n.dtrans[i][j] != F {
				ntransations += 1
			}
		}

		//只输出包含有效跳转的节点对应的行,例如节点1就没有任何跳转，因此就不输出它对应的行
		if ntransations > 0 {
			fmt.Fprintf(n.fp, "%s%d, ", name, i)
		} else {
			fmt.Fprintf(n.fp, "NULL ,")
		}

		nprinted -= 1
		if nprinted <= 0 {
			fmt.Fprintf(n.fp, "\n        ")
			nprinted = 10
		}
	}

	fmt.Fprintf(n.fp, "%s%d\n};\n\n", name, i)

	//结束输出第二部分
	return numCells
}

func (n *NfaDfaConverter) pnext(name string) {
	/*
		这里输出一段c语言代码，它们的作用是根据输入字符，查询上面生成的压缩跳转表后，跳转到下一个状态节点
	*/

	toptext := []string{"unsigned int c ;", "int cur_state;", "{",
		"/*给定当前状态和输入字符，返回跳转后的下一个状态节点*/",
	}
	boptext := []string{
		"    int i;",
		"",
		"    if ( p ) ",
		"    {",
		"        if ((i = *p++) == 0)",
		"            return p[ c ];",
		"",
		"        for(; --i >= 0; p += 2)",
		"            if( c == p[0]  )",
		"                return p[1];",
		"    }",
		"    return unsigned char (-1);",
		"}",
	}

	fmt.Fprintf(n.fp, "\n/*------------------------------------*/\n")
	fmt.Fprintf(n.fp, "%s %s yynext(cur_state, c)\n", SCLASS, TYPE)
	for _, str := range toptext {
		fmt.Fprintf(n.fp, "%s\n", str)
	}
	fmt.Fprintf(n.fp, "    %s    *p = %s[cur_state];\n", TYPE, name)
	for _, str := range boptext {
		fmt.Fprintf(n.fp, "%s\n", str)
	}
}

func (n *NfaDfaConverter) DoPairCompression() {
	n.pairs(DTRAN_NAME, 10, 1)

	n.pnext(DTRAN_NAME)
}

func (n *NfaDfaConverter) getUnMarked() *DFA {
	for ; n.lastMarked < n.nstates; n.lastMarked++ {
		debug := 0
		if n.dstates[n.lastMarked].state == 5 {
			debug = 1
			fmt.Printf("debug: %d", debug)
		}
		if n.dstates[n.lastMarked].mark == false {
			return &n.dstates[n.lastMarked]
		}
	}

	return nil
}

func (n *NfaDfaConverter) compareNfaSlice(setOne []*NFA, setTwo []*NFA) bool {
	//比较两个集合的元素是否相同
	if len(setOne) != len(setTwo) {
		return false
	}

	equal := false
	for _, nfaOne := range setOne {
		for _, nfaTwo := range setTwo {
			if nfaTwo == nfaOne {
				equal = true
				break
			}
		}

		if equal != true {
			return false
		}
	}

	return true
}

func (n *NfaDfaConverter) hasDfaContainsNfa(nfaSet []*NFA) (bool, int) {
	//查看是否存在dfa节点它对应的nfa节点集合与输入的集合相同
	for _, dfa := range n.dstates {
		if n.compareNfaSlice(dfa.set, nfaSet) == true {
			return true, dfa.state
		}
	}

	return false, -1
}

func (n *NfaDfaConverter) addDfaState(epsilonResult *EpsilonResult) int {
	//根据当前nfa节点集合构造一个新的dfa节点
	nextState := F
	if n.nstates >= DFA_MAX {
		panic("Too many DFA states")
	}

	nextState = n.nstates
	n.nstates += 1
	n.dstates[nextState].set = epsilonResult.results
	n.dstates[nextState].mark = false
	n.dstates[nextState].acceptString = epsilonResult.acceptStr
	//该节点是否为终结节点
	n.dstates[nextState].isAccepted = epsilonResult.hasAccepted

	n.dstates[nextState].anchor = epsilonResult.anchor
	n.dstates[nextState].state = nextState //记录当前dfa节点的编号s

	n.printDFAState(&n.dstates[nextState])
	fmt.Print("\n")

	return nextState
}

func (n *NfaDfaConverter) printDFAState(dfa *DFA) {
	fmt.Printf("DFA state : %d, it is nfa are: {", dfa.state)
	for _, nfa := range dfa.set {
		fmt.Printf("%d,", nfa.state)
	}

	fmt.Printf("}")
}

func (n *NfaDfaConverter) MakeDTran(start *NFA) {
	//根据输入的nfa状态机起始节点构造dfa状态机的跳转表
	startStates := make([]*NFA, 0)
	startStates = append(startStates, start)
	statesCopied := make([]*NFA, len(startStates))
	copy(statesCopied, startStates)

	//先根据起始状态的求Epsilon闭包操作的结果，由此获得第一个dfa节点
	epsilonResult := EpsilonClosure(statesCopied)
	n.dstates[0].set = epsilonResult.results
	n.dstates[0].anchor = epsilonResult.anchor
	n.dstates[0].acceptString = epsilonResult.acceptStr
	n.dstates[0].mark = false

	//debug purpose
	n.printDFAState(&n.dstates[0])
	fmt.Print("\n")
	nextState := 0
	n.nstates = 1 //当前已经有一个dfa节点
	//先获得第一个没有设置其跳转边的dfa节点
	current := n.getUnMarked()
	for current != nil {
		current.mark = true
		for c := 0; c < MAX_CHARS; c++ {
			nfaSet := move(current.set, c)
			if len(nfaSet) > 0 {
				statesCopied = make([]*NFA, len(nfaSet))
				copy(statesCopied, nfaSet)
				epsilonResult = EpsilonClosure(statesCopied)
				nfaSet = epsilonResult.results
			}

			if len(nfaSet) == 0 {
				nextState = F
			} else {
				//如果当前没有那个dfa节点对应的nfa节点集合和当前nfaSet相同，那么就增加一个新的dfa节点
				isExist, state := n.hasDfaContainsNfa(nfaSet)
				if isExist == false {
					nextState = n.addDfaState(epsilonResult)
				} else {
					nextState = state
				}
			}

			//设置dfa跳转表
			n.dtrans[current.state][c] = nextState
		}

		current = n.getUnMarked()
	}
}

func (n *NfaDfaConverter) PrintDfaTransition() {
	for i := 0; i < DFA_MAX; i++ {
		if n.dstates[i].mark == false {
			break
		}

		for j := 0; j < MAX_CHARS; j++ {
			if n.dtrans[i][j] != F {
				n.printDFAState(&n.dstates[i])
				fmt.Print(" jump to : ")
				n.printDFAState(&n.dstates[n.dtrans[i][j]])
				fmt.Printf("by character %s\n", string(j))
			}
		}
	}
}

func (n *NfaDfaConverter) initGroups() {
	//先把节点根据接收状态分为两个分区
	for i := 0; i < n.nstates; i++ {
		if n.dstates[i].isAccepted {
			n.groups[1] = append(n.groups[1], n.dstates[i].state)
			//记录状态点对应的分区
			n.inGroups[n.dstates[i].state] = 1
		} else {
			n.groups[0] = append(n.groups[0], n.dstates[i].state)
			n.inGroups[n.dstates[i].state] = 0
		}
	}

	n.numGroups = 2
}

func (n *NfaDfaConverter) printGroups() {
	//打印当前分区的信息
	for i := 0; i < n.numGroups; i++ {
		group := n.groups[i]
		fmt.Printf("分区号: %d", i)
		fmt.Println("分区节点如下:")
		for j := 0; j < len(group); j++ {
			fmt.Printf("%d ", group[j])
		}
		fmt.Printf("\n")
	}
}

func (n *NfaDfaConverter) minimizeGroups() {
	for {
		oldNumGroups := n.numGroups
		for current := 0; current < n.numGroups; current++ {
			//遍历每个分区，将不属于同一个分区的节点拿出来形成新的分区
			if len(n.groups[current]) <= 1 {
				//对于只有1个元素的分区不做处理
				continue
			}

			idx := 0
			//获取分区第一个元素
			first := n.groups[current][idx]
			newPartition := false
			for idx+1 < len(n.groups[current]) {
				next := n.groups[current][idx+1]
				//如果分区还有未处理的元素，那么看其是否跟first对应元素属于同一分区
				for c := MAX_CHARS - 1; c >= 0; c-- {
					gotoFirst := n.dtrans[first][c]
					gotoNext := n.dtrans[next][c]
					if gotoFirst != gotoNext && (gotoFirst == F || gotoNext == F || n.inGroups[gotoFirst] != n.inGroups[gotoNext]) {
						//如果first和next对应的两个节点在接收相同输入后跳转的节点不在同一分区，那么需要将next对应节点加入新分区
						//先将next对应节点从当前分区拿走
						n.groups[current] = append(n.groups[current][:idx+1], n.groups[current][idx+2:]...)
						n.groups[n.numGroups] = append(n.groups[n.numGroups], next)
						n.inGroups[next] = n.numGroups
						newPartition = true
						break
					}
				}

				if !newPartition {
					//如果next没有被拿出当前分区，那么idx要增加去指向下一个元素
					idx += 1
				} else {
					//如果next被挪出当前分区，那么idx不用变就能指向下一个元素♀️
					newPartition = false
				}
			}

			if len(n.groups[n.numGroups]) > 0 {
				//有新的分区生成，因此分区计数要加1
				n.numGroups += 1
			}
		}

		if oldNumGroups == n.numGroups {
			//如果没有新分区生成，算法结束
			break
		}
	}

	n.nstates = n.numGroups
	n.printGroups()
}

func (n *NfaDfaConverter) fixTran() {
	newDTran := make([][]int, DFA_MAX)
	//新建一个跳转表
	for i := 0; i < n.numGroups; i++ {
		newDTran[i] = make([]int, MAX_CHARS)
	}

	/*
		我们把当前分区号对应一个新的DFA节点，当前分区(用A表示)中取出一个节点，根据输入字符c获得其跳转的节点。
		然后根据跳转节点获得其所在分区(用B表示)，那么我们就得到新节点A在接收字符c后跳转到B节点	。
		这里我们从当前分区取出一个节点就行，因为在minimizeGroups中我们已经确保最终的分区中，里面每个节点在接收
		同样的字符后，所跳转的节点所在的分区肯定是一样的。
	*/
	for i := 0; i < n.numGroups; i++ {
		//从当前分区取出一个节点即可
		state := n.groups[i][0]
		for c := MAX_CHARS - 1; c >= 0; c-- {
			if n.dtrans[state][c] == F {
				newDTran[i][c] = F
			} else {
				destState := n.dtrans[state][c]
				destPartition := n.inGroups[destState]
				newDTran[i][c] = destPartition
			}
		}
	}

	n.dtrans = newDTran
}

func (n *NfaDfaConverter) MinimizeDFA() {
	n.initGroups()
	n.minimizeGroups()
	n.fixTran()
}

func (n *NfaDfaConverter) PrintMinimizeDFATran() {
	for i := 0; i < n.numGroups; i++ {
		for j := 0; j < MAX_CHARS; j++ {
			if n.dtrans[i][j] != F {
				fmt.Printf("from state %d jump to state %d with input: %s\n", i, n.dtrans[i][j], string(j))
			}
		}
	}
}
