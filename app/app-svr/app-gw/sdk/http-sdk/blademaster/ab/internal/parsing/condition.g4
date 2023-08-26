grammar condition;

/*
Parser
*/
condition
    : NOT condition                     #logicalNot
    | condition AND condition           #logicalAnd
    | condition OR condition            #logicalOr
    | LPAREN condition RPAREN           #paren
    | compare                           #logicalOp2
    | inOrNotIn                         #inOrNotInOp2
    | variableDeclarator                #commomOp
    ;

compare
    : variableDeclarator op variableDeclarator #logicalOp
    ;

inOrNotIn
    : variableDeclarator (IN | NIN) LPAREN variableDeclarator (',' variableDeclarator)* RPAREN   #inOrNotInOp
    ;

variableDeclarator
    : TRUE              #logicalTrue
    | FALSE             #logicalFalse
    | INT               #commonInt
    | DOUBLE            #commonDouble
    | STRING            #commonString
    | IDENTIFIER        #variable
    ;

/*
Lexer
*/
op : GE | LE | EQ | NE | GT | LT;

EQ : '==';
NE : '!=';
GE : '>=';
LE : '<=';
GT : '>' ;
LT : '<' ;

OR : '||';
AND: '&&';
NOT: '!';

IN : 'in' ;
NIN: 'not in' ;

LPAREN: '(';
RPAREN: ')';
QUOTE : '"';
COMMA : ',';

TRUE  : 'true' ;
FALSE : 'false';
INT : [0-9]+;                              // 整数
DOUBLE : [1-9][0-9]*|[0]|([0-9]+[.][0-9]+);// 小数
STRING : '"' ('\\"'|.)*? '"' ;             // 字符串

IDENTIFIER: Letter LetterOrDigit*;

fragment LetterOrDigit
    : Letter
    | [0-9]
    ;
fragment Letter
    : [a-zA-Z$_]
    ;

WS : [ \r\n\t] + -> skip;