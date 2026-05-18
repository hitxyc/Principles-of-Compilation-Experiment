package main

import (
	"fmt"
	"strconv"
	"strings"
)

type GrammaticalAnalysis struct {
	productions   []Production
	nonTerminal   map[string]bool
	terminalTable map[string]bool
	actionTable   []map[string]ActionEntry
	gotoTable     []map[string]GotoEntry
	firstCollect  map[string]map[string]bool
}

func NewGrammaticalAnalysis() *GrammaticalAnalysis {
	return &GrammaticalAnalysis{}
}

type ActionEntry struct {
	Action string // "S"=Shift, "R"=Reduce, "ACC"=Accept, "ERR"=Error
	Value  int    // Shift到的状态或Reduce的产生式编号
}

type GotoEntry struct {
	Value int // Goto到的状态
}

type Production struct {
	LHS string
	RHS []string
}

type item struct {
	prod      Production
	dot       int
	lookAhead string
	prodIndex int
}

type state struct {
	items []item
}

func itemKey(it item) string {
	return it.prod.LHS + "->" + strings.Join(it.prod.RHS, " ") + "|" + strconv.Itoa(it.dot) + "@" + it.lookAhead
}

func symbolAfterDot(it item) string {
	if it.dot < len(it.prod.RHS) {
		return it.prod.RHS[it.dot]
	}
	return ""
}

func (analysis *GrammaticalAnalysis) isNonTerminal(sym string) bool {
	return analysis.nonTerminal[sym]
}

func (analysis *GrammaticalAnalysis) initFirst() {
	analysis.firstCollect = make(map[string]map[string]bool)
	// 初始化
	for k := range analysis.nonTerminal {
		if _, ok := analysis.firstCollect[k]; !ok {
			analysis.firstCollect[k] = make(map[string]bool)
		}
	}
	for k := range analysis.terminalTable {
		if _, ok := analysis.firstCollect[k]; !ok {
			analysis.firstCollect[k] = make(map[string]bool)
			analysis.firstCollect[k][k] = true
		}
	}

	changed := true
	for changed {
		changed = false
		for _, p := range analysis.productions {
			LHS := p.LHS
			RHS := p.RHS
			canBeEpsilon := true

			for i := 0; i < len(RHS) && canBeEpsilon; i++ {
				sym := RHS[i]
				canBeEpsilon = false
				if !analysis.nonTerminal[sym] {
					if !analysis.firstCollect[LHS][sym] {
						analysis.firstCollect[LHS][sym] = true
						changed = true
					}
					break
				}
				for k := range analysis.firstCollect[sym] {
					if k == "ε" {
						canBeEpsilon = true
					} else {
						if !analysis.firstCollect[LHS][k] {
							analysis.firstCollect[LHS][k] = true
							changed = true
						}
					}
				}
			}
			if canBeEpsilon {
				if !analysis.firstCollect[LHS]["ε"] {
					analysis.firstCollect[LHS]["ε"] = true
					changed = true
				}
			}
		}
	}
}

func (analysis *GrammaticalAnalysis) firstOfSequence(strS []string) []string {
	firstCollect := make(map[string]bool)
	isNullable := true
	for _, sym := range strS {
		for t := range analysis.firstCollect[sym] {
			if t != "ε" {
				firstCollect[t] = true
			}
		}
		if !analysis.firstCollect[sym]["ε"] {
			isNullable = false
			break
		}
	}
	if isNullable {
		firstCollect["ε"] = true
	}
	res := make([]string, 0)
	for k := range firstCollect {
		res = append(res, k)
	}
	return res
}

func (analysis *GrammaticalAnalysis) closure(st state, productions []Production) state {
	set := make(map[string]struct{})
	for _, it := range st.items {
		set[itemKey(it)] = struct{}{}
	}
	for i := 0; i < len(st.items); i++ {
		it := st.items[i]
		// 如果 dot 前是非终结符，加它的产生式
		if sym := symbolAfterDot(it); analysis.isNonTerminal(sym) {
			for i, p := range productions {
				if p.LHS == sym {
					beta := it.prod.RHS[it.dot+1:]
					betaPlusLookahead := append(beta, it.lookAhead)
					firstOfNewItem := analysis.firstOfSequence(betaPlusLookahead) // 计算 FIRST(beta a)
					for _, f := range firstOfNewItem {
						newItem := item{p, 0, f, i}
						if _, ok := set[itemKey(newItem)]; !ok {
							st.items = append(st.items, newItem)
							set[itemKey(newItem)] = struct{}{}
						}
					}
				}
			}
		}
	}
	return st
}

func allSymbols(curr state) []string {
	set := make(map[string]struct{})
	for _, it := range curr.items {
		sym := symbolAfterDot(it)
		if sym != "" {
			set[sym] = struct{}{}
		}
	}
	symbols := make([]string, 0)
	for sym := range set {
		symbols = append(symbols, sym)
	}
	return symbols
}

func (analysis *GrammaticalAnalysis) gotoFunc(curr state, sym string, productions []Production) state {
	newState := state{}
	for _, it := range curr.items {
		if it.dot < len(it.prod.RHS) && it.prod.RHS[it.dot] == sym {
			newState.items = append(newState.items, item{it.prod, it.dot + 1, it.lookAhead, it.prodIndex})
		}
	}
	newState = analysis.closure(newState, productions)
	return newState
}

func indexOfState(states []state, newState state) int {
	for i, st := range states {
		if len(st.items) != len(newState.items) {
			continue
		}
		set := make(map[string]struct{})
		for _, it := range st.items {
			set[itemKey(it)] = struct{}{}
		}
		flag := true
		for _, it := range newState.items {
			if _, ok := set[itemKey(it)]; !ok {
				flag = false
				break
			}
		}
		if flag {
			return i
		}
	}
	return -1
}

func (analysis *GrammaticalAnalysis) init(grammar string) {
	// 1.解析文法 && 初始化
	lines := strings.Split(grammar, "\n")
	productions := make([]Production, 0)
	for _, line := range lines {
		HS := strings.Split(line, "->")
		lhs := strings.TrimSpace(HS[0])
		rhs := strings.Fields(strings.TrimSpace(HS[1]))
		if len(rhs) == 1 && rhs[0] == "ε" {
			rhs = []string{} // 空串
		}
		productions = append(productions, Production{lhs, rhs})
	}
	analysis.productions = productions
	// 得到所有终结符与非终结符
	nonTerminals := make(map[string]bool)
	for _, it := range productions {
		nonTerminals[it.LHS] = true
	}
	analysis.nonTerminal = nonTerminals
	allTerminals := make(map[string]bool)
	for _, p := range productions {
		for _, sym := range p.RHS {
			if !analysis.isNonTerminal(sym) { // 非终结符 → 终结符
				allTerminals[sym] = true
			}
		}
	}
	allTerminals["$"] = true
	analysis.terminalTable = allTerminals
	// 得到first集
	analysis.initFirst()
	// 2.初始化LR(0)状态
	startState := state{}
	startState.items = append(startState.items, item{prod: productions[0], dot: 0, lookAhead: "$"})
	states := []state{startState}

	// 3.构造DFA（闭包 + GOTO）
	queue := []int{0}
	transitions := make(map[int]map[string]int)
	for len(queue) > 0 {
		idx := queue[0]
		queue = queue[1:]
		curr := states[idx]
		// 闭包运算
		curr = analysis.closure(curr, productions)
		// 保存闭包后的状态（用于后续状态比较与查询）
		states[idx] = curr
		// GOTO操作
		symbols := allSymbols(curr)
		for _, sym := range symbols {
			newState := analysis.gotoFunc(curr, sym, productions)
			j := indexOfState(states, newState)
			if j == -1 {
				states = append(states, newState)
				j = len(states) - 1
				queue = append(queue, j)
			}
			if transitions[idx] == nil {
				transitions[idx] = make(map[string]int)
			}
			transitions[idx][sym] = j
		}
	}
	// 4.构造action表和goto表
	actionTable := make([]map[string]ActionEntry, len(states))
	for i := 0; i < len(states); i++ {
		actionTable[i] = make(map[string]ActionEntry)
	}
	gotoTable := make([]map[string]GotoEntry, len(states))
	for i := 0; i < len(states); i++ {
		gotoTable[i] = make(map[string]GotoEntry)
	}
	for i, st := range states {
		for _, it := range st.items {
			// 点在最后，表示可规约
			if it.dot == len(it.prod.RHS) {
				lhs := it.prod.LHS
				// 特殊情况：起始产生式规约 -> Accept（仅当 lookahead 为 $）
				if lhs == "S'" && it.lookAhead == "$" {
					actionTable[i]["$"] = ActionEntry{Action: "ACC"}
				} else {
					// 仅在该项的 lookahead 上填 R（Reduce）
					t := it.lookAhead
					if _, ok := actionTable[i][t]; !ok { // 没有冲突
						actionTable[i][t] = ActionEntry{Action: "R", Value: it.prodIndex}
					} else {
						fmt.Printf("LR(1) 冲突在状态 %d, 符号 %s: %+v (existing) vs prod %d (new)\n", i, t, actionTable[i][t], it.prodIndex)
					}
				}
				continue
			}
			// 移进操作
			sym := symbolAfterDot(it)
			if analysis.isNonTerminal(sym) {
				if j, ok := transitions[i][sym]; ok {
					gotoTable[i][sym] = GotoEntry{Value: j}
				}
			} else {
				if j, ok := transitions[i][sym]; ok {
					actionTable[i][sym] = ActionEntry{Action: "S", Value: j}
				}
			}
		}
	}
	analysis.actionTable = actionTable
	analysis.gotoTable = gotoTable
}

type StackItem struct {
	State  int
	Symbol string
}

func (analysis *GrammaticalAnalysis) analyze(tokens []Token) bool {
	stack := []StackItem{{State: 0, Symbol: ""}}
	index := 0
	for {
		topState := stack[len(stack)-1].State
		var a string
		if index < len(tokens) {
			if tokens[index].Type == IDENTIFIER {
				a = "id"
			} else if tokens[index].Type == CONSTANT {
				a = "num"
			} else {
				a = tokens[index].Lexeme
			}
		} else {
			a = "$"
		}
		entry, ok := analysis.actionTable[topState][a]
		if !ok {
			entry.Action = "ERR"
		}
		switch entry.Action {
		case "S":
			stack = append(stack, StackItem{State: entry.Value, Symbol: a})
			index++
			fmt.Printf("Shift in: %v\n", a)
		case "R":
			prod := analysis.productions[entry.Value]
			for i := 0; i < len(prod.RHS); i++ {
				stack = stack[:len(stack)-1]
			}
			t := stack[len(stack)-1].State
			stack = append(stack, StackItem{State: analysis.gotoTable[t][prod.LHS].Value, Symbol: prod.LHS})
			fmt.Printf("Reduce by: %s -> %v\n", prod.LHS, prod.RHS)
		case "ACC":
			fmt.Println("Syntax Accepted!")
			return true
		default:
			fmt.Printf("Syntax Error at token %d: %s\n", index, a)
			return false
		}
	}
}
