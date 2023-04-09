package main

import (
	"fmt"
	"nfa"
)

func main() {
	lexReader, _ := nfa.NewLexReader("input.lex", "output.py")
	lexReader.Head()
	parser, _ := nfa.NewRegParser(lexReader)
	start := parser.Parse()
	parser.PrintNFA(start)
	//str := "3.14"
	//if nfa.NfaMatchString(start, str) {
	//	fmt.Printf("string %s is accepted by given regular expression\n", str)
	//}
	nfaConverter := nfa.NewNfaDfaConverter()
	nfaConverter.MakeDTran(start)
	nfaConverter.PrintDfaTransition()

	nfaConverter.MinimizeDFA()
	fmt.Println("---------new DFA transition table ----")
	nfaConverter.PrintMinimizeDFATran()

	//nfaConverter.DoPairCompression()
	nfaConverter.DoSquash()
}
