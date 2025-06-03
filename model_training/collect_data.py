"""
Usage: python3 collect_data.py <target_file> <save_path>
Finds and saves all comments in a code file for fine-tuning a SLM
"""
import tree_sitter_python as tspython
import tree_sitter_go as tsgo
import tree_sitter_lua as tslua
from tree_sitter import Language, Parser
import sys
import os

PYTHON_LANGUAGE = Language(tspython.language())
GO_LANGUAGE = Language(tsgo.language())
LUA_LANGUAGE = Language(tslua.language())

PYTHON_QUERY = PYTHON_LANGUAGE.query("""
(function_definition
 name: (identifier) @function.def
 body: (block) @function.block
 )
""")

LUA_QUERY = PYTHON_LANGUAGE.query("""
(function_definition)
""")

GO_QUERY = PYTHON_LANGUAGE.query("""
(function_definition)
""")


def main(args):
    file_path = args[1]
    save_path = args[2]
    lang = None
    query = None
    if file_path.endswith(".py"):
        lang = PYTHON_LANGUAGE
        query = PYTHON_QUERY
    elif file_path.endswith(".lua"):
        lang = LUA_LANGUAGE
        query = LUA_QUERY
    elif file_path.endswith(".go"):
        lang = GO_LANGUAGE
        query = GO_QUERY
    else:
        print("Error language not supported!")
        return 1
    raw_code = bytes("", "utf-8")
    with open(file_path, "rb") as f:
        raw_code = bytes(f.read())
    # do a treesitter query
    parser = Parser(lang)
    tree = parser.parse(raw_code, )

    print(
        query.captures(tree.root_node)
    )
    return 0


if __name__ == "__main__":
    if len(sys.argv) < 3:
        print(__doc__)
    else:
        main(sys.argv)
