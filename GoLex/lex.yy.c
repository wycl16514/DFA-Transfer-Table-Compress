
/*--------------------------------------
* The Yy_cmap[] and Yy_rmap arrays are used as follows:
* 
*  next_state= Yydtran[ Yy_rmap[current_state] ][ Yy_cmap[input_char] ];
* 
* Character positions in the Yy_cmap array are:
* 
*    ^@  ^A  ^B  ^C  ^D  ^E  ^F  ^G  ^H  ^I  ^J  ^K  ^L  ^M  ^N  ^O
*    ^P  ^Q  ^R  ^S  ^T  ^U  ^V  ^W  ^X  ^Y  ^Z  ^[  ^\  ^]  ^^  ^_
*         !   "   #   $   %   &   '   (   )   *   +   ,   -   .   /
*     0   1   2   3   4   5   6   7   8   9   :   ;   <   =   >   ?
*     @   A   B   C   D   E   F   G   H   I   J   K   L   M   N   O
*     P   Q   R   S   T   U   V   W   X   Y   Z   [   \   ]   ^   _
*     `   a   b   c   d   e   f   g   h   i   j   k   l   m   n   o
*     p   q   r   s   t   u   v   w   x   y   z   {   |   }   ~   DEL
*/

static unsigned char Yy_cmap[128]=
{
  0,0,0,0,0,0,0,0,0,0,
    0,0,0,0,0,0,0,0,0,0,
    0,0,0,0,0,0,0,0,0,0,
    0,0,0,0,0,0,0,0,0,0,
    0,0,0,0,0,0,1,0,2,2,
    2,2,2,2,2,2,2,2,0,0,
    0,0,0,0,0,0,0,0,0,0,
    0,0,0,0,0,0,0,0,0,0,
    0,0,0,0,0,0,0,0,0,0,
    0,0,0,0,0,0,0,0,0,0,
    0,0,0,0,0,0,0,0,0,0,
    0,0,0,0,0,0,0,0,0,0,
    0,0,0,0,0,0,0,0
};

static  unsigned char  Yy_rmap[6]
{
    0,1,2,3,4,5
};

static unsigned char Yy_nxt[6][3]=
{
/*  0 */ {-1, 2, 4},
/*  1 */ {-1, -1, -1},
/*  2 */ {-1, -1, 1},
/*  3 */ {-1, -1, 3},
/*  4 */ {-1, 3, 5},
/*  5 */ {-1, 2, 5} 
};

/*--------------------------------------
* yy_next(state,c) is given the current state number and input
* character and evaluates to the next state.
*/

define yy_next(state, c) (Yy_nxt[Yy_rmap[state]][Yy_cmap[c]])
