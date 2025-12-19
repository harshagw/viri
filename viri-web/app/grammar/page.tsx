import Link from "next/link";

export default function GrammarPage() {
  return (
    <>
      <main className="flex-1">
        {/* Header */}
        <section className="py-16 border-b border-border">
          <div className="container mx-auto px-6 max-w-4xl">
            <h1 className="text-5xl font-bold mb-4 font-mono text-primary">grammar</h1>
            <p className="text-muted-foreground">EBNF notation. Click any rule to navigate.</p>
          </div>
        </section>

        <div className="container mx-auto px-6 py-12 max-w-4xl">
          {/* EBNF Legend */}
          <section className="mb-12 pb-8 border-b border-border">
            <p className="text-sm text-muted-foreground mb-4 uppercase tracking-wider">notation</p>
            <div className="grid grid-cols-2 md:grid-cols-3 gap-4 text-sm">
              <div>
                <code className="font-mono text-primary">|</code> <span className="text-muted-foreground">alternative</span>
              </div>
              <div>
                <code className="font-mono text-primary">[ ]</code> <span className="text-muted-foreground">optional</span>
              </div>
              <div>
                <code className="font-mono text-primary">{"{ }"}</code> <span className="text-muted-foreground">repetition</span>
              </div>
            </div>
          </section>

          {/* Grammar Rules */}
          <div className="space-y-6">
            {/* Program Structure */}
            <p className="text-sm text-muted-foreground uppercase tracking-wider pt-4">program structure</p>

            <GrammarRule
              id="program"
              name="program"
              definition={
                <>
                  {"{"} <RuleLink href="#import">import</RuleLink> {"}"} {"{"} <RuleLink href="#declaration">declaration</RuleLink> {"}"} <Lexical>EOF</Lexical>
                </>
              }
              referencedBy={[]}
            />

            <GrammarRule
              id="import"
              name="import"
              definition={
                <>
                  <Token>import</Token> <Lexical>STRING</Lexical> <Token>as</Token> <Lexical>IDENTIFIER</Lexical> <Token>;</Token>
                </>
              }
              referencedBy={["program"]}
            />

            {/* Declarations */}
            <p className="text-sm text-muted-foreground uppercase tracking-wider pt-8">declarations</p>

            <GrammarRule
              id="declaration"
              name="declaration"
              definition={
                <>
                  <RuleLink href="#classDecl">classDecl</RuleLink> | <RuleLink href="#funDecl">funDecl</RuleLink> | <RuleLink href="#varDecl">varDecl</RuleLink>{" "}
                  | <RuleLink href="#constDecl">constDecl</RuleLink> | <RuleLink href="#statement">statement</RuleLink>
                </>
              }
              referencedBy={["program", "block"]}
            />

            <GrammarRule
              id="varDecl"
              name="varDecl"
              definition={
                <>
                  [ <Token>export</Token> ] <Token>var</Token> <Lexical>IDENTIFIER</Lexical> [ <Token>=</Token>{" "}
                  <RuleLink href="#expression">expression</RuleLink> ] <Token>;</Token>
                </>
              }
              referencedBy={["declaration", "forStmt"]}
            />

            <GrammarRule
              id="constDecl"
              name="constDecl"
              definition={
                <>
                  [ <Token>export</Token> ] <Token>const</Token> <Lexical>IDENTIFIER</Lexical> <Token>=</Token>{" "}
                  <RuleLink href="#expression">expression</RuleLink> <Token>;</Token>
                </>
              }
              referencedBy={["declaration", "forStmt"]}
            />

            <GrammarRule
              id="funDecl"
              name="funDecl"
              definition={
                <>
                  [ <Token>export</Token> ] <Token>fun</Token> <RuleLink href="#function">function</RuleLink>
                </>
              }
              referencedBy={["declaration"]}
            />

            <GrammarRule
              id="classDecl"
              name="classDecl"
              definition={
                <>
                  [ <Token>export</Token> ] <Token>class</Token> <Lexical>IDENTIFIER</Lexical> [ <Token>&lt;</Token> <Lexical>IDENTIFIER</Lexical> ]{" "}
                  <Token>{"{"}</Token> {"{"} <RuleLink href="#function">function</RuleLink> {"}"} <Token>{"}"}</Token>
                </>
              }
              referencedBy={["declaration"]}
            />

            <GrammarRule
              id="function"
              name="function"
              definition={
                <>
                  <Lexical>IDENTIFIER</Lexical> <Token>(</Token> [ <RuleLink href="#parameters">parameters</RuleLink> ] <Token>)</Token>{" "}
                  <RuleLink href="#block">block</RuleLink>
                </>
              }
              referencedBy={["funDecl", "classDecl"]}
            />

            <GrammarRule
              id="parameters"
              name="parameters"
              definition={
                <>
                  <Lexical>IDENTIFIER</Lexical> {"{"} <Token>,</Token> <Lexical>IDENTIFIER</Lexical> {"}"}
                </>
              }
              referencedBy={["function", "functionExpr"]}
            />

            {/* Statements */}
            <p className="text-sm text-muted-foreground uppercase tracking-wider pt-8">statements</p>

            <GrammarRule
              id="statement"
              name="statement"
              definition={
                <>
                  <RuleLink href="#exprStmt">exprStmt</RuleLink> | <RuleLink href="#forStmt">forStmt</RuleLink> | <RuleLink href="#ifStmt">ifStmt</RuleLink> |{" "}
                  <RuleLink href="#printStmt">printStmt</RuleLink> | <RuleLink href="#returnStmt">returnStmt</RuleLink> |{" "}
                  <RuleLink href="#whileStmt">whileStmt</RuleLink> | <RuleLink href="#breakStmt">breakStmt</RuleLink> |{" "}
                  <RuleLink href="#continueStmt">continueStmt</RuleLink> | <RuleLink href="#block">block</RuleLink>
                </>
              }
              referencedBy={["declaration", "forStmt", "ifStmt", "whileStmt"]}
            />

            <GrammarRule
              id="block"
              name="block"
              definition={
                <>
                  <Token>{"{"}</Token> {"{"} <RuleLink href="#declaration">declaration</RuleLink> {"}"} <Token>{"}"}</Token>
                </>
              }
              referencedBy={["function", "functionExpr", "statement"]}
            />

            <GrammarRule
              id="exprStmt"
              name="exprStmt"
              definition={
                <>
                  <RuleLink href="#expression">expression</RuleLink> <Token>;</Token>
                </>
              }
              referencedBy={["statement", "forStmt"]}
            />

            <GrammarRule
              id="printStmt"
              name="printStmt"
              definition={
                <>
                  <Token>print</Token> <RuleLink href="#expression">expression</RuleLink> <Token>;</Token>
                </>
              }
              referencedBy={["statement"]}
            />

            <GrammarRule
              id="returnStmt"
              name="returnStmt"
              definition={
                <>
                  <Token>return</Token> [ <RuleLink href="#expression">expression</RuleLink> ] <Token>;</Token>
                </>
              }
              referencedBy={["statement"]}
            />

            <GrammarRule
              id="ifStmt"
              name="ifStmt"
              definition={
                <>
                  <Token>if</Token> <Token>(</Token> <RuleLink href="#expression">expression</RuleLink> <Token>)</Token>{" "}
                  <RuleLink href="#statement">statement</RuleLink> [ <Token>else</Token> <RuleLink href="#statement">statement</RuleLink> ]
                </>
              }
              referencedBy={["statement"]}
            />

            <GrammarRule
              id="whileStmt"
              name="whileStmt"
              definition={
                <>
                  <Token>while</Token> <Token>(</Token> <RuleLink href="#expression">expression</RuleLink> <Token>)</Token>{" "}
                  <RuleLink href="#statement">statement</RuleLink>
                </>
              }
              referencedBy={["statement"]}
            />

            <GrammarRule
              id="forStmt"
              name="forStmt"
              definition={
                <>
                  <Token>for</Token> <Token>(</Token> ( <RuleLink href="#varDecl">varDecl</RuleLink> | <RuleLink href="#constDecl">constDecl</RuleLink> |{" "}
                  <RuleLink href="#exprStmt">exprStmt</RuleLink> | <Token>;</Token> ) [ <RuleLink href="#expression">expression</RuleLink> ] <Token>;</Token> [{" "}
                  <RuleLink href="#expression">expression</RuleLink> ] <Token>)</Token> <RuleLink href="#statement">statement</RuleLink>
                </>
              }
              referencedBy={["statement"]}
            />

            <GrammarRule
              id="breakStmt"
              name="breakStmt"
              definition={
                <>
                  <Token>break</Token> <Token>;</Token>
                </>
              }
              referencedBy={["statement"]}
            />

            <GrammarRule
              id="continueStmt"
              name="continueStmt"
              definition={
                <>
                  <Token>continue</Token> <Token>;</Token>
                </>
              }
              referencedBy={["statement"]}
            />

            {/* Expressions */}
            <p className="text-sm text-muted-foreground uppercase tracking-wider pt-8">expressions</p>

            <GrammarRule
              id="expression"
              name="expression"
              definition={
                <>
                  <RuleLink href="#assignment">assignment</RuleLink>
                </>
              }
              referencedBy={[
                "varDecl",
                "constDecl",
                "exprStmt",
                "forStmt",
                "ifStmt",
                "whileStmt",
                "printStmt",
                "returnStmt",
                "call",
                "arrayLiteral",
                "hashEntry",
              ]}
            />

            <GrammarRule
              id="assignment"
              name="assignment"
              definition={
                <>
                  ( <RuleLink href="#call">call</RuleLink> <Token>.</Token> <Lexical>IDENTIFIER</Lexical> | <RuleLink href="#call">call</RuleLink>{" "}
                  <Token>[</Token> <RuleLink href="#expression">expression</RuleLink> <Token>]</Token> | <Lexical>IDENTIFIER</Lexical> ) <Token>=</Token>{" "}
                  <RuleLink href="#assignment">assignment</RuleLink> | <RuleLink href="#logic_or">logic_or</RuleLink>
                </>
              }
              referencedBy={["expression"]}
            />

            <GrammarRule
              id="logic_or"
              name="logic_or"
              definition={
                <>
                  <RuleLink href="#logic_and">logic_and</RuleLink> {"{"} <Token>or</Token> <RuleLink href="#logic_and">logic_and</RuleLink> {"}"}
                </>
              }
              referencedBy={["assignment"]}
            />

            <GrammarRule
              id="logic_and"
              name="logic_and"
              definition={
                <>
                  <RuleLink href="#equality">equality</RuleLink> {"{"} <Token>and</Token> <RuleLink href="#equality">equality</RuleLink> {"}"}
                </>
              }
              referencedBy={["logic_or"]}
            />

            <GrammarRule
              id="equality"
              name="equality"
              definition={
                <>
                  <RuleLink href="#comparison">comparison</RuleLink> {"{"} ( <Token>!=</Token> | <Token>==</Token> ){" "}
                  <RuleLink href="#comparison">comparison</RuleLink> {"}"}
                </>
              }
              referencedBy={["logic_and"]}
            />

            <GrammarRule
              id="comparison"
              name="comparison"
              definition={
                <>
                  <RuleLink href="#term">term</RuleLink> {"{"} ( <Token>&gt;</Token> | <Token>&gt;=</Token> | <Token>&lt;</Token> | <Token>&lt;=</Token> ){" "}
                  <RuleLink href="#term">term</RuleLink> {"}"}
                </>
              }
              referencedBy={["equality"]}
            />

            <GrammarRule
              id="term"
              name="term"
              definition={
                <>
                  <RuleLink href="#factor">factor</RuleLink> {"{"} ( <Token>-</Token> | <Token>+</Token> ) <RuleLink href="#factor">factor</RuleLink> {"}"}
                </>
              }
              referencedBy={["comparison"]}
            />

            <GrammarRule
              id="factor"
              name="factor"
              definition={
                <>
                  <RuleLink href="#unary">unary</RuleLink> {"{"} ( <Token>/</Token> | <Token>*</Token> ) <RuleLink href="#unary">unary</RuleLink> {"}"}
                </>
              }
              referencedBy={["term"]}
            />

            <GrammarRule
              id="unary"
              name="unary"
              definition={
                <>
                  ( <Token>!</Token> | <Token>-</Token> ) <RuleLink href="#unary">unary</RuleLink> | <RuleLink href="#call">call</RuleLink>
                </>
              }
              referencedBy={["factor"]}
            />

            <GrammarRule
              id="call"
              name="call"
              definition={
                <>
                  <RuleLink href="#primary">primary</RuleLink> {"{"} <Token>(</Token> [ <RuleLink href="#arguments">arguments</RuleLink> ] <Token>)</Token> |{" "}
                  <Token>.</Token> <Lexical>IDENTIFIER</Lexical> | <Token>[</Token> <RuleLink href="#expression">expression</RuleLink> <Token>]</Token> {"}"}
                </>
              }
              referencedBy={["unary", "assignment"]}
            />

            <GrammarRule
              id="primary"
              name="primary"
              definition={
                <>
                  <Token>true</Token> | <Token>false</Token> | <Token>nil</Token> | <Token>this</Token> | <Lexical>NUMBER</Lexical> | <Lexical>STRING</Lexical>{" "}
                  | <Lexical>IDENTIFIER</Lexical> | <Token>(</Token> <RuleLink href="#expression">expression</RuleLink> <Token>)</Token> | <Token>super</Token>{" "}
                  <Token>.</Token> <Lexical>IDENTIFIER</Lexical> | <RuleLink href="#arrayLiteral">arrayLiteral</RuleLink> |{" "}
                  <RuleLink href="#hashLiteral">hashLiteral</RuleLink> | <RuleLink href="#functionExpr">functionExpr</RuleLink>
                </>
              }
              referencedBy={["call"]}
            />

            <GrammarRule
              id="arguments"
              name="arguments"
              definition={
                <>
                  <RuleLink href="#expression">expression</RuleLink> {"{"} <Token>,</Token> <RuleLink href="#expression">expression</RuleLink> {"}"}
                </>
              }
              referencedBy={["call"]}
            />

            {/* Literals */}
            <p className="text-sm text-muted-foreground uppercase tracking-wider pt-8">literals</p>

            <GrammarRule
              id="arrayLiteral"
              name="arrayLiteral"
              definition={
                <>
                  <Token>[</Token> [ <RuleLink href="#expression">expression</RuleLink> {"{"} <Token>,</Token>{" "}
                  <RuleLink href="#expression">expression</RuleLink> {"}"} ] <Token>]</Token>
                </>
              }
              referencedBy={["primary"]}
            />

            <GrammarRule
              id="hashLiteral"
              name="hashLiteral"
              definition={
                <>
                  <Token>{"{"}</Token> [ <RuleLink href="#hashEntry">hashEntry</RuleLink> {"{"} <Token>,</Token>{" "}
                  <RuleLink href="#hashEntry">hashEntry</RuleLink> {"}"} ] <Token>{"}"}</Token>
                </>
              }
              referencedBy={["primary"]}
            />

            <GrammarRule
              id="hashEntry"
              name="hashEntry"
              definition={
                <>
                  <RuleLink href="#expression">expression</RuleLink> <Token>:</Token> <RuleLink href="#expression">expression</RuleLink>
                </>
              }
              referencedBy={["hashLiteral"]}
            />

            <GrammarRule
              id="functionExpr"
              name="functionExpr"
              definition={
                <>
                  <Token>fun</Token> <Token>(</Token> [ <RuleLink href="#parameters">parameters</RuleLink> ] <Token>)</Token>{" "}
                  <RuleLink href="#block">block</RuleLink>
                </>
              }
              referencedBy={["primary"]}
            />
          </div>

          {/* Footer note */}
          <div className="mt-16 pt-8 border-t border-border">
            <p className="text-sm text-muted-foreground">
              <Lexical>IDENTIFIER</Lexical>, <Lexical>NUMBER</Lexical>, <Lexical>STRING</Lexical>, <Lexical>EOF</Lexical> are lexical tokens.
            </p>
          </div>
        </div>
      </main>

      {/* Footer */}
      <footer className="border-t border-border py-8">
        <div className="container mx-auto px-6 max-w-4xl">
          <div className="flex justify-between items-center text-sm text-muted-foreground">
            <Link href="/" className="hover:text-foreground">
              ← home
            </Link>
            <span>viri grammar</span>
          </div>
        </div>
      </footer>
    </>
  );
}

function GrammarRule({ id, name, definition, referencedBy }: { id: string; name: string; definition: React.ReactNode; referencedBy: string[] }) {
  return (
    <div id={id} className="scroll-mt-20">
      <div className="flex items-baseline gap-4 mb-2">
        <Link href={`#${id}`} className="text-muted-foreground hover:text-foreground">
          #
        </Link>
        <h3 className="font-mono font-semibold text-primary">{name}</h3>
      </div>
      <div className="pl-8 mb-2">
        <code className="font-mono text-sm bg-muted px-3 py-2 inline-block border border-border">{definition}</code>
      </div>
      {referencedBy.length > 0 && (
        <div className="pl-8 text-sm text-muted-foreground">
          referenced by:{" "}
          {referencedBy.map((ref, i) => (
            <span key={ref}>
              {i > 0 && ", "}
              <RuleLink href={`#${ref}`}>{ref}</RuleLink>
            </span>
          ))}
        </div>
      )}
    </div>
  );
}

function RuleLink({ href, children }: { href: string; children: React.ReactNode }) {
  return (
    <Link href={href} className="text-primary hover:underline">
      {children}
    </Link>
  );
}

function Token({ children }: { children: React.ReactNode }) {
  return <span className="text-green-600 dark:text-green-400">&quot;{children}&quot;</span>;
}

function Lexical({ children }: { children: React.ReactNode }) {
  return <span className="text-amber-600 dark:text-amber-400">{children}</span>;
}
