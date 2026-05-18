package main

import "fmt"

func main() {
	lexicalAnalyzer := NewLexicalAnalysis()
	lexicalAnalyzer.Analyze("int a;\nint b;\nint c;\na=2;\nb=1;\n\t\t  if (a>b)\nc=a+b;\nelse\nc=a-b;")
	grammaticalAnalyzer := NewGrammaticalAnalysis()
	grammaticalAnalyzer.init(`S' -> L
       L  -> L S
       L  -> S
       S  -> D
       S  -> A
       S  -> I
       D  -> int id ;
       A  -> id = E ;
       I  -> if ( B ) S X
       X  -> else S
       X  -> ε
       B  -> B || C
       B  -> C
       C  -> C && R
       C  -> R
       R  -> E relop E
       R  -> E
       relop -> >
       relop -> <
       relop -> >=
       relop -> <=
       relop -> ==
       E  -> E + T
       E  -> E - T
       E  -> T
       T  -> T * F
       T  -> T / F
       T  -> F
       F  -> ( E )
       F  -> id
       F  -> num`)
	grammaticalAnalyzer.analyze(lexicalAnalyzer.TokenTable)
	fmt.Println("hello lexicalAnalyzer")
}
