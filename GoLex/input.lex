%{
   int FCON = 1;
   int ICON = 2;
%}
D  [0-9]
%%
({D}*\.{D}|{D}\.{D}*)  {printf("%s is a float number", yytext); return FCON;}
%%