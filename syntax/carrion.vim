" Vim syntax file for Carrion programming language
" Language: Carrion
" Maintainer: Carrion LSP Team
" Latest Revision: 2025

if exists("b:current_syntax")
  finish
endif

" Keywords
syn keyword carrionSpellKeyword spell nextgroup=carrionFunctionName skipwhite
syn keyword carrionClassKeyword grim nextgroup=carrionClassName skipwhite
syn keyword carrionKeyword arcane arcanespell init self super
syn keyword carrionConditional if otherwise else
syn keyword carrionLoop for while in
syn keyword carrionControl skip stop return
syn keyword carrionException attempt ensnare resolve raise check
syn keyword carrionOperatorWord and or not
syn keyword carrionBoolean True False
syn keyword carrionConstant None
syn keyword carrionImport import as global
syn keyword carrionMatch match case
syn keyword carrionOther var ignore main autoclose

" Function and class names
syn match carrionFunctionName "\w\+" contained
syn match carrionClassName "\w\+" contained

" Comments
syn match carrionComment "#.*$"
syn region carrionBlockComment start="/\*" end="\*/" contains=carrionTodo
syn region carrionTripleComment start="```" end="```" contains=carrionTodo
syn keyword carrionTodo contained TODO FIXME XXX NOTE

" Strings
syn region carrionString start='"' skip='\\"' end='"' contains=carrionEscape
syn region carrionString start="'" skip="\\'" end="'" contains=carrionEscape
syn region carrionFString start='f"' skip='\\"' end='"' contains=carrionEscape,carrionInterpolation
syn region carrionFString start="f'" skip="\\'" end="'" contains=carrionEscape,carrionInterpolation

" String escape sequences
syn match carrionEscape contained "\\[nrtbf\\\"']"
syn match carrionEscape contained "\\u[0-9a-fA-F]\{4}"
syn match carrionEscape contained "\\x[0-9a-fA-F]\{2}"

" F-string interpolation
syn region carrionInterpolation contained start="{" end="}" contains=carrionNumber,carrionFloat,carrionString,carrionKeyword

" Numbers
syn match carrionNumber "\<\d\+\>"
syn match carrionFloat "\<\d\+\.\d\+\>"
syn match carrionFloat "\<\d\+\.\?\d*[eE][+-]\?\d\+\>"
syn match carrionBinary "\<0[bB][01]\+\>"
syn match carrionOctal "\<0[oO][0-7]\+\>"
syn match carrionHex "\<0[xX][0-9a-fA-F]\+\>"

" Operators
syn match carrionOperator "\v\+\+|--|\+=|-=|\*=|/=|%=|\*\*="
syn match carrionOperator "\v\+|-|\*|/|%|\*\*"
syn match carrionOperator "\v\=\=|!\=|\<\=|\>\=|\<|\>"
syn match carrionOperator "\v\="
syn match carrionOperator "\v\&\&|\|\|"

" Delimiters
syn match carrionDelimiter "\v[\[\](){},;:.]"

" Special identifiers
syn match carrionSpecial "\<__\w*__\>"

" Error highlighting for invalid syntax
syn match carrionError "\v[^\x00-\x7f]" " Non-ASCII characters (if not in strings)

" Define highlight groups
hi def link carrionSpellKeyword Keyword
hi def link carrionClassKeyword Keyword
hi def link carrionKeyword Keyword
hi def link carrionConditional Conditional
hi def link carrionLoop Repeat
hi def link carrionControl Statement
hi def link carrionException Exception
hi def link carrionOperatorWord Operator
hi def link carrionBoolean Boolean
hi def link carrionConstant Constant
hi def link carrionImport Include
hi def link carrionMatch Statement
hi def link carrionOther Keyword

hi def link carrionFunctionName Function
hi def link carrionClassName Type

hi def link carrionComment Comment
hi def link carrionBlockComment Comment
hi def link carrionTripleComment Comment
hi def link carrionTodo Todo

hi def link carrionString String
hi def link carrionFString String
hi def link carrionEscape Special
hi def link carrionInterpolation Special

hi def link carrionNumber Number
hi def link carrionFloat Float
hi def link carrionBinary Number
hi def link carrionOctal Number
hi def link carrionHex Number

hi def link carrionOperator Operator
hi def link carrionDelimiter Delimiter
hi def link carrionSpecial Special
hi def link carrionError Error

let b:current_syntax = "carrion"