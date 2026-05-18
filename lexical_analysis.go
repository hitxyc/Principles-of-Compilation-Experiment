package main

type TokenType int

const (
	KEYWORD TokenType = iota
	IDENTIFIER
	CONSTANT
	OPERATOR
	SEPARATOR
	ERROR
)

type Token struct {
	Type   TokenType
	Lexeme string
	LineNo int
}

type Symbol struct {
	Type   TokenType
	Name   string
	LineNo int
}

type LexicalAnalyzer struct {
	KeywordTable map[string]bool
	TokenTable   []Token
	SymbolTable  []Symbol
}

func NewLexicalAnalysis() *LexicalAnalyzer {
	KeyWordTable := map[string]bool{"int": true, "if": true, "else": true, "return": true, "for": true, "while": true}
	return &LexicalAnalyzer{
		KeywordTable: KeyWordTable,
		TokenTable:   []Token{},
		SymbolTable:  []Symbol{},
	}
}

func (analyzer *LexicalAnalyzer) isKeyword(word string) bool {
	return analyzer.KeywordTable[word]
}

func (analyzer *LexicalAnalyzer) isOperator(c byte) bool {
	return c == '+' || c == '-' || c == '*' || c == '/' || c == '=' || c == '<' || c == '>'
}

func (analyzer *LexicalAnalyzer) isSeparator(c byte) bool {
	return c == '(' || c == ')' || c == '{' || c == '}' || c == ',' || c == ';'
}

func (analyzer *LexicalAnalyzer) isWhitespace(c byte) bool {
	return c == ' ' || c == '\t' || c == '\n'
}

func (analyzer *LexicalAnalyzer) isDigitalOrLetter(c byte) bool {
	return (c >= '0' && c <= '9') || (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || c == '_'
}

func (analyzer *LexicalAnalyzer) addToken(typ TokenType, lexeme string, lineNo int) {
	analyzer.TokenTable = append(analyzer.TokenTable, Token{typ, lexeme, lineNo})
}

func (analyzer *LexicalAnalyzer) addSymbol(typ TokenType, name string, lineNo int) {
	analyzer.SymbolTable = append(analyzer.SymbolTable, Symbol{typ, name, lineNo})
}

func (analyzer *LexicalAnalyzer) addWord(word string, lineNo int) {
	if analyzer.isKeyword(word) {
		analyzer.addToken(KEYWORD, word, lineNo)
	} else {
		if word[0] >= '0' && word[0] <= '9' {
			analyzer.addToken(CONSTANT, word, lineNo)
			analyzer.addSymbol(CONSTANT, word, lineNo)
		} else {
			analyzer.addToken(IDENTIFIER, word, lineNo)
			analyzer.addSymbol(IDENTIFIER, word, lineNo)
		}
	}
}

func (analyzer *LexicalAnalyzer) Analyze(str string) {
	pointer, lineNo := 0, 0
	n := len(str)
	for pointer < n {
		if analyzer.isSeparator(str[pointer]) {
			analyzer.addToken(SEPARATOR, str[pointer:pointer+1], lineNo)
		} else if analyzer.isOperator(str[pointer]) {
			operator := ""
			operator, pointer = analyzer.scanOperator(str, pointer)
			analyzer.addToken(OPERATOR, operator, lineNo)
			if operator == "" {
				analyzer.addToken(ERROR, str[pointer:pointer+1], lineNo)
			}
		} else if analyzer.isWhitespace(str[pointer]) {
			if str[pointer] == '\n' {
				lineNo++
			}
		} else if analyzer.isDigitalOrLetter(str[pointer]) {
			word := ""
			word, pointer = analyzer.scanWord(str, pointer)
			analyzer.addWord(word, lineNo)
		} else {
			analyzer.addToken(ERROR, str[pointer:pointer+1], lineNo)
		}
		pointer++
	}
}

func (analyzer *LexicalAnalyzer) scanOperator(str string, pointer int) (string, int) {
	const (
		INIT = iota
		PLUS_STATE
		MINUS_STATE
		EQ_STATE
		LT_STATE
		GT_STATE
	)

	state := INIT
	temp := ""
	for pointer < len(str) {
		c := str[pointer]
		switch state {
		case INIT:
			switch c {
			case '+':
				state = PLUS_STATE
				temp = "+"
				pointer++
			case '-':
				state = MINUS_STATE
				temp = "-"
				pointer++
			case '=':
				state = EQ_STATE
				temp = "="
				pointer++
			case '<':
				state = LT_STATE
				temp = "<"
				pointer++
			case '>':
				state = GT_STATE
				temp = ">"
				pointer++
			case '*', '/', '%': // 单字符运算符
				return string(c), pointer
			default:
				return "", pointer // 非运算符
			}
		// 超前搜索
		case PLUS_STATE:
			if pointer < len(str) {
				next := str[pointer]
				if next == '+' {
					return "++", pointer
				} else if next == '=' {
					return "+=", pointer
				}
			}
			return "+", pointer - 1

		case MINUS_STATE:
			if pointer < len(str) {
				next := str[pointer]
				if next == '-' {
					return "--", pointer
				} else if next == '=' {
					return "-=", pointer
				}
			}
			return "-", pointer - 1

		case EQ_STATE:
			if pointer < len(str) && str[pointer] == '=' {
				return "==", pointer
			}
			return "=", pointer - 1

		case LT_STATE:
			if pointer < len(str) && str[pointer] == '=' {
				return "<=", pointer
			}
			return "<", pointer - 1

		case GT_STATE:
			if pointer < len(str) && str[pointer] == '=' {
				return ">=", pointer
			}
			return ">", pointer - 1
		}
	}
	return temp, pointer
}

func (analyzer *LexicalAnalyzer) scanWord(str string, pointer int) (string, int) {
	start := pointer
	for pointer < len(str) {
		c := str[pointer]
		if (c >= '0' && c <= '9') ||
			(c >= 'a' && c <= 'z') ||
			(c >= 'A' && c <= 'Z') ||
			c == '_' {
			pointer++
		} else {
			break
		}
	}
	return str[start:pointer], pointer - 1
}
