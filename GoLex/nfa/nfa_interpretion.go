package nfa

import (
	"fmt"
	"math"
)

type EpsilonResult struct {
	/*
		如果结果集合中包含终结点，那么accept_str对应终结点的accept字符串，anchor对于终结点的Anchor对象
	*/
	results     []*NFA
	acceptStr   string
	hasAccepted bool
	anchor      Anchor
}

func stackContains(stack []*NFA, elem *NFA) bool {
	for _, i := range stack {
		if i == elem {
			return true
		}
	}

	return false
}

func EpsilonClosure(input []*NFA) *EpsilonResult {
	acceptState := math.MaxInt
	result := &EpsilonResult{}

	for len(input) > 0 {
		node := input[len(input)-1]
		input = input[0 : len(input)-1]
		//epsilon-closure的操作结果一定包含输入节点集合
		result.results = append(result.results, node)
		/*
			如果有多个终结节点，那么选取状态值最小的那个作为接收点
		*/
		if node.next == nil && node.state < acceptState {
			result.acceptStr = node.accept
			result.anchor = node.anchor
			result.hasAccepted = true
		}

		if node.edge == EPSILON {
			if node.next != nil && stackContains(input, node.next) == false {
				input = append(input, node.next)
			}

			if node.next2 != nil && stackContains(input, node.next2) == false {
				input = append(input, node.next2)
			}
		}
	}

	return result
}

func move(input []*NFA, c int) []*NFA {
	result := make([]*NFA, 0)
	for _, elem := range input {
		if int(elem.edge) == c || (elem.edge == CCL && elem.bitset[string(c)] == true) {
			result = append(result, elem.next)
		}
	}

	return result
}

func printEpsilonClosure(input []*NFA, output []*NFA) {
	fmt.Printf("%s({", "epsilon-closure")
	for _, elem := range input {
		fmt.Printf("%d,", elem.state)
	}
	fmt.Printf("})={")
	for _, elem := range output {
		fmt.Printf("%d, ", elem.state)
	}
	fmt.Printf("})\n")
}

func printMove(input []*NFA, output []*NFA, c string) {
	fmt.Printf("move({")
	for _, elem := range input {
		fmt.Printf("%d,", elem.state)
	}
	fmt.Printf("}, %s)={", c)
	for _, elem := range output {
		fmt.Printf("%d, ", elem.state)
	}
	fmt.Printf("})\n")
}

func NfaMatchString(state *NFA, str string) bool {
	/*
		state是NFA状态机的起始节点，str对应要匹配的字符串
	*/
	startStates := make([]*NFA, 0)
	startStates = append(startStates, state)
	statesCopied := make([]*NFA, len(startStates))
	copy(statesCopied, startStates)
	result := EpsilonClosure(statesCopied)

	printEpsilonClosure(startStates, result.results)

	strRead := ""
	strAccepted := false
	for i, char := range str {
		moveResult := move(result.results, int(char))
		printMove(result.results, moveResult, string(char))
		if moveResult == nil {
			fmt.Printf("%s is not accepted by nfa machine\n", str)
		}
		strRead += string(char)
		statesCopied = make([]*NFA, len(moveResult))
		copy(statesCopied, moveResult)
		result = EpsilonClosure(moveResult)
		printEpsilonClosure(statesCopied, result.results)
		if result.hasAccepted {
			fmt.Printf("current string : %s is accepted by the machine\n", strRead)
		}

		if i == len(str)-1 {
			strAccepted = result.hasAccepted
		}
	}

	return strAccepted
}
