import { useRef, useState, useEffect } from 'react'
import { FloppyDiskIcon, FileArchiveIcon, MinusIcon, PlusIcon, CopyIcon } from '@phosphor-icons/react'
import JSZip from 'jszip';
import { CodeBlock } from './components/CodeBlock';
import { SyntaxEditor } from './components/SyntaxEditor';
import { AIBlockly } from './components/AIBlockly';
import './App.css'

declare function mistToXml(mistCode: string, componentDefinitions: Record<string, string[]>): string;

function AnchorHeading({ level, id, children }: { level: 'h2' | 'h3' | 'h4', id: string, children: React.ReactNode }) {
  const Tag = level;
  const handleClick = () => {
    window.history.pushState({}, '', `#${id}`);
    document.getElementById(id)?.scrollIntoView({ behavior: 'smooth' });
  };

  return (
    <Tag id={id} className="anchor-heading">
      {children}
      <a href={`#${id}`} className="anchor-link" onClick={(e) => { e.preventDefault(); handleClick(); }} aria-label="Link to this section">
        #
      </a>
    </Tag>
  );
}

function App() {
  const demoRef = useRef<HTMLDivElement>(null);
  const docsRef = useRef<HTMLDivElement>(null);
  const [code, setCode] = useState(`// Type Falcon Code`);
  const [error, setError] = useState<string | null>("Syntax Error: Expected '}' at line 3, column 5");

  const scrollToDemo = () => {
    demoRef.current?.scrollIntoView({ behavior: 'smooth' });
  };

  const scrollToDocs = () => {
    docsRef.current?.scrollIntoView({ behavior: 'smooth' });
  };

  /**
   * Exports all top-level blocks in the main workspace as PNGs.
   */
  function exportAllBlocks() {
    // Get the main Blockly workspace.
    const workspace = (window as any).Blockly?.getMainWorkspace();
    if (!workspace) {
      console.error("Main workspace not found.");
      return;
    }

    // Get all the top-level blocks on the workspace.
    const topBlocks = workspace.getTopBlocks(true);

    if (topBlocks.length === 0) {
      console.log("No blocks to export.");
      return;
    }

    // Iterate over each top-level block and trigger the export.
    // This will download a PNG for each top block and its children.
    topBlocks.forEach((block: any) => {
      (window as any).Blockly?.exportBlockAsPng(block);
    });
  }

  const handleZip = async () => {
    const workspace = (window as any).Blockly?.getMainWorkspace();
    if (!workspace) return;

    const topBlocks = workspace.getTopBlocks(true);
    if (topBlocks.length === 0) return;

    const zip = new JSZip();
    const promises = topBlocks.map((block: any) => {
      return new Promise<void>((resolve) => {
        // Use the requested blockToPngBlob function
        (window as any).Blockly.blockToPngBlob(block).then((blob: Blob) => {
          // Name the file based on the block type or id, ensure unique names if needed
          // For simplicity using Block Type and ID
          const filename = `${block.type}_${block.id}.png`;
          zip.file(filename, blob);
          resolve();
        }).catch((err: any) => {
          console.error("Error generating png for block", block, err);
          resolve(); // Resolve anyway to continue zip generation
        });
      });
    });

    try {
      await Promise.all(promises);
      const content = await zip.generateAsync({ type: "blob" });

      const url = URL.createObjectURL(content);
      const a = document.createElement('a');
      a.href = url;
      a.download = 'falcon_blocks.zip';
      document.body.appendChild(a);
      a.click();
      document.body.removeChild(a);
      URL.revokeObjectURL(url);
    } catch (error) {
      console.error("Error creating zip", error);
    }
  };

  const handleCodeChange = (newCode: string) => {
    setCode(newCode);
    translateToBlocks(newCode);
    setError(null);
  };

  const handleZoomIn = () => {
    const workspace = (window as any).Blockly?.getMainWorkspace();
    workspace?.zoom(0, 0, 1);
  };

  const handleZoomOut = () => {
    const workspace = (window as any).Blockly?.getMainWorkspace();
    workspace?.zoom(0, 0, -1);
  };


  // Mist -> XML
  function translateToBlocks(mistCode: string) {
    try {
      const xmlCode = mistToXml(mistCode, {});
      console.log("Generated XML Code:", xmlCode);
      renderBlocks(xmlCode);
    } catch (error) {
      console.log(error)
    }
  }

  function renderBlocks(xmlGenerated: string) {
    // First clear the existing workspace, and inject the new blocks
    const xmlStrings = xmlGenerated.split("\u0000");
    const workspace = (window as any).Blockly?.getMainWorkspace();
    workspace.clear()

    const blocks = [];

    for (let i = 0; i < xmlStrings.length; i++) {
      const xmlString = xmlStrings[i].trim();
      if (!xmlString || xmlString.replace(/\0/g, '').trim() === '') {
        continue;
      }

      console.log(xmlString);
      const xml = (window as any).Blockly?.utils.xml.textToDom(xmlString);
      const xmlBlock = xml.firstElementChild;
      const block = (window as any).Blockly?.Xml.domToBlock(xmlBlock, workspace);
      block.initSvg(); // Init all blocks first
      blocks.push(block); // Save for rendering later
    }

    for (const block of blocks) {
      workspace.requestRender(block);
    }
    sortBlocks();
  }

  function sortBlocks() {
    // Sort all the blocks in order
    const item = (window as any).Blockly?.ContextMenuRegistry.registry.getItem("appinventor_arrange_vertical");

    if (item && typeof item.callback === "function") {
      const workspace = (window as any).Blockly?.getMainWorkspace();

      const fakeScope = { workspace: workspace, };
      item.callback(fakeScope, null);
    } else {
      console.error("Callback not found or item is invalid");
    }
  };


  // Handle initial hash navigation
  useEffect(() => {
    const hash = window.location.hash.slice(1);
    if (hash) {
      setTimeout(() => {
        const element = document.getElementById(hash);
        element?.scrollIntoView({ behavior: 'smooth' });
      }, 100);
    }
  }, []);

  return (
    <div className="app-container">
      <section className="hero-section" id="home">
        <h1>Falcon</h1>
        <p>
          Syntax-Block transformation engine for App Inventor.
        </p>
        <div className="info-card">
          <a href="#documentation" onClick={(e) => { e.preventDefault(); scrollToDocs(); }}>Documentation</a>
          <a href="https://github.com/XomaDev/Falcon">GitHub</a>
        </div>
        <div className="scroll-indicator" onClick={scrollToDemo}>
          <svg width="30" height="30" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
            <path d="M7 13L12 18L17 13M7 6L12 11L17 6" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" />
          </svg>
        </div>
      </section>

      <section className="demo-section" ref={demoRef} id="try-it-out">
        <h2>Try it out</h2>
        <div className="playground-container">
          <div className="editor-card">
            <div className="card-header">
              <h3>Input</h3>
            </div>
            <SyntaxEditor
              value={code}
              onChange={handleCodeChange}
              placeholder="Type Falcon syntax here..."
            />
            {error && (
              <div className="editor-error">
                <span>{error}</span>
                <button
                  onClick={() => navigator.clipboard.writeText(error)}
                  className="error-copy-btn"
                  title="Copy Error"
                  aria-label="Copy Error"
                >
                  <CopyIcon size={16} />
                </button>
              </div>
            )}
          </div>
          <div className="preview-card">
            <div className="card-header">
              <h3>Blocks Preview</h3>
              <div style={{ display: 'flex', gap: '0.5rem' }}>
                <button onClick={handleZoomIn} className="icon-btn" title="Zoom In">
                  <PlusIcon size={20} />
                </button>
                <button onClick={handleZoomOut} className="icon-btn" title="Zoom Out">
                  <MinusIcon size={20} />
                </button>
                <button onClick={exportAllBlocks} className="icon-btn" title="Save Blocks">
                  <FloppyDiskIcon size={20} />
                </button>
                <button onClick={handleZip} className="icon-btn" title="Zip Blocks">
                  <FileArchiveIcon size={20} />
                </button>
              </div>
            </div>
            <AIBlockly />
          </div>
        </div>
      </section>

      <section className="docs-section" ref={docsRef} id="documentation">
        <div className="docs-container">
          <div className="docs-content">
            <AnchorHeading level="h2" id="documentation">Documentation</AnchorHeading>
            <p>
              Falcon is a language designed for App Inventor to enable syntax-based programming and for incorporating agenting coding abilities.
            </p>

            <AnchorHeading level="h3" id="quirks">Quirks</AnchorHeading>
            <ul>
              <li>Falcon follows 1-based indexing.</li>
              <li>Falcon variables are dynamically typed. Do not declare variables.</li>
              <li>Lists and dictionaries are passed as references.</li>
              <li>Falcon follows Kotlin's style of functional expressions.</li>
              <li>Falcon does not have a return statement; the last expression in a body is returned.</li>
              <li>Falcon does NOT have a try-catch or a throw statement.</li>
              <li>Only single-line comments using double slash <code>//</code> are supported.</li>
              <li>Do not use <code>_</code> in place of unused variables.</li>
              <li>Variables can never be uninitialized.</li>
              <li>Always keep the last expression for returning functions.</li>
              <li>If the contents of two strings are numeric, then they can be numerically operated on, e.g. <code>"2" + "3.14"</code> is valid code.</li>
            </ul>

            <AnchorHeading level="h3" id="data-types">Data types</AnchorHeading>
            <ul>
              <li>String: <code>"Hello, world!"</code></li>
              <li>Boolean: <code>true</code> and <code>false</code></li>
              <li>Number: <code>123</code> and <code>3.14</code></li>
              <li>List: <code>[1, 2, 3, 4]</code></li>
              <li>Dictionary: <code>{`{"Animal": "Tiger", "Scientific Name": "Panthera tigris"}`}</code></li>
              <li>Colour: <code>#FFFFFF</code></li>
            </ul>

            <AnchorHeading level="h3" id="operators">Operators</AnchorHeading>
            <ul>
              <li>Arithmetic: <code>+, -, *, /, % (remainder), ^ (power)</code></li>
              <li>Logical: <code>&&, ||</code></li>
              <li>Bitwise: <code>&, |, ~ (xor)</code></li>
              <li>Equality: <code>==, !=</code></li>
              <li>Relational: <code>&lt;, &lt;=, &gt;, &gt;=</code></li>
              <li>Text lexicographic: <code>===</code> (text equals), <code>!==</code> (text not equals), <code>&lt;&lt;</code> (text less than), <code>&gt;&gt;</code> (text greater than)</li>
              <li>Unary: <code>!</code> (not), <code>-</code> (negate)</li>
              <li>Join: <code>"Hello " _ "World!"</code></li>
              <li>Pair: <code>"Fruit": "Mango"</code></li>
              <li>Question (<code>?</code>):
                <ul>
                  <li>To check a value for a specific type (text, number, list, dict)<br />E.g. <code>"Hello" ? text</code></li>
                  <li>Check for a number type (number, base10, hexa, bin)<br />E.g. <code>"1010" ? bin</code> is a true expression.</li>
                  <li>Check for empty text (emptyText) or an empty list (emptyList)<br />E.g. <code>[] ? emptyList</code> or <code>"Cat" ? emptyText</code></li>
                </ul>
              </li>
            </ul>

            <AnchorHeading level="h3" id="operator-precedence">Operator precedence</AnchorHeading>
            <p>The precedence of an operator dictates its parse order. E.g. <code>*</code> and <code>/</code> are parsed before <code>+</code> and <code>-</code>. It is similar to that of Java. Below is a ranking from the lowest to the highest precedence.</p>
            <ul>
              <li>Assignment <code>=</code></li>
              <li>Pair <code>:</code></li>
              <li>TextJoin <code>_</code></li>
              <li>LogicOr <code>||</code></li>
              <li>LogicAnd <code>&&</code></li>
              <li>BitwiseOr <code>|</code></li>
              <li>BitwiseAnd <code>&</code></li>
              <li>BitwiseXor <code>~</code></li>
              <li>Equality <code>==, !=, ===, !==</code></li>
              <li>Relational <code>&lt;, &lt;=, &gt;, &gt;=, &lt;&lt;, &gt;&gt;</code></li>
              <li>Binary <code>+, -</code></li>
              <li>BinaryL1 <code>*, /, %</code></li>
              <li>BinaryL2 <code>^</code></li>
            </ul>

            <AnchorHeading level="h3" id="variables">Variables</AnchorHeading>
            <h4>Global variable</h4>
            <p>A global variable is always declared at the root:</p>
            <CodeBlock code={`global name = "Kumaraswamy B G"

// access the global variable
println(this.name)`} />

            <h4>Local variable</h4>
            <CodeBlock code={`local age = 17

// access the local variable
println(age)`} />

            <AnchorHeading level="h3" id="if-else">If else</AnchorHeading>
            <p>If-else can be a statement or an expression depending on the context.</p>
            <CodeBlock code={`local x = 8
local y = 12

if (x > y) {
  println("X is greater")
} else if (y > x) {
  println("Y is greater")
} else {
  println("They both are equal!")
}`} />
            <p>Used as an expression:</p>
            <CodeBlock code={`println(  if (x > y) "X is greater" else if  (y > x) "Y is greater" else "They both are equal!"  )`} />

            <AnchorHeading level="h3" id="while-loop">While loop</AnchorHeading>
            <CodeBlock code={`local x = 8

while (true) {
  x = x + 1
  if (x == 5) {
    break  
  }
}`} />

            <AnchorHeading level="h3" id="for-n-loop">For n loop</AnchorHeading>
            <CodeBlock code={`for (i: 1 .. 10 step 2) {
  println(i)
}`} />
            <p>The step clause is optional and defaults to 1.</p>

            <AnchorHeading level="h3" id="each-loop">Each loop</AnchorHeading>
            <p>To iterate over a list:</p>
            <CodeBlock code={`local names = ["India", "Japan", "Russia", "Germany"]

for (country in names) {
  println(country)
}`} />
            <p>Or over a dictionary:</p>
            <CodeBlock code={`local animalInfo = { "Animal": "Tiger", "Scientific Name": "Panthera tigris" }

for (key, value in animalDetail) {
  println(key _ " : " _ value) // e.g prints  "Animal: Tiger" to the console
}`} />

            <AnchorHeading level="h3" id="functions">Functions</AnchorHeading>
            <p>Functions are declared using the <code>func</code> keyword.</p>
            <h4>Void function</h4>
            <CodeBlock code={`func fooBar(x, y) {
  println(x + y)
}`} />
            <h4>Result function</h4>
            <p>Use the <code>=</code> symbol followed by an expression between curly braces.</p>
            <CodeBlock code={`func double(n) = { n * 2 }`} />
            <p>Or multiple expressions:</p>
            <CodeBlock code={`func FibSum(n) = {
  if (n < 2) {
    n  
  } else {
    FibSum(n - 1) + FibSum(n - 2)
  }
}`} />
            <p>Note that there is no return statement in Falcon. The last statement in a body is taken as the output of an expression.</p>

            <AnchorHeading level="h3" id="math">Math</AnchorHeading>
            <ul>
              <li><code>dec(string), bin(string), octal(string), hexa(string)</code> - Parse a static constant string from the respective base.</li>
              <li><code>sqrt(number), abs(number), neg(number), log(number), exp(number)</code></li>
              <li><code>round(number), ceil(number), floor(number)</code></li>
              <li><code>sin(number), cos(number), tan(number)</code></li>
              <li><code>asin(number), acos(number), atan(number), atan2(a, b)</code></li>
              <li><code>degrees(number), radians(number)</code></li>
              <li><code>decToHex(number), decToBin(number), hexToDec(number), binToDec(number)</code></li>
              <li><code>randInt(from, to), randFloat()</code></li>
              <li><code>setRandSeed(number)</code> - sets the random generator seed</li>
              <li><code>min(...), max(...)</code></li>
              <li><code>avgOf(list), maxOf(list), minOf(list), geoMeanOf(), stdDevOf(), stdErrOf(), modeOf(list)</code></li>
              <li><code>mod(x, y), rem(x, y), quot(x, y)</code></li>
              <li><code>formatDecimal(number, places)</code></li>
            </ul>

            <AnchorHeading level="h3" id="control">Control</AnchorHeading>
            <ul>
              <li><code>println(any)</code></li>
              <li><code>openScreen(name)</code></li>
              <li><code>openScreenWithValue(name, value)</code></li>
              <li><code>closeScreenWithValue(value)</code></li>
              <li><code>getStartValue()</code></li>
              <li><code>closeScreen()</code></li>
              <li><code>closeApp()</code></li>
              <li><code>getPlainStartText()</code></li>
            </ul>

            <AnchorHeading level="h3" id="values">Values</AnchorHeading>
            <ul>
              <li><code>copyList(list), copyDict(dict)</code></li>
              <li><code>makeColor(rgb_list), splitColor(number)</code></li>
            </ul>

            <AnchorHeading level="h3" id="methods">Methods</AnchorHeading>
            <p>Example: <code>"Hello  ".trim()</code></p>
            <h4>Text</h4>
            <ul>
              <li><code>textLen(), trim(), uppercase(), lowercase(), reverse()</code></li>
              <li><code>startsWith(piece), contains(piece), containsAny(list), containsAll(list)</code></li>
              <li><code>split(at), splitAtFirst(at), splitAtAny(list), splitAtFirstOfAny(list), splitAtSpaces()</code></li>
              <li><code>csvRowToList(), csvTableToList()</code></li>
              <li><code>segment(from, length), replace(target, replacement)</code></li>
              <li><code>replaceFrom(map), replaceFromLongestFirst(map)</code></li>
            </ul>
            <h4>List</h4>
            <ul>
              <li><code>listLen(), add(items...), containsItem(item), indexOf(item)</code></li>
              <li><code>insert(index, item), remove(index), appendList(list)</code></li>
              <li><code>lookupInPairs(key, notfound), join(separator), slice(i1, i2)</code></li>
              <li><code>random(), reverseList(), sort()</code></li>
              <li><code>toCsvRow(), toCsvTable()</code></li>
              <li><code>allButFirst(), allButLast()</code></li>
              <li><code>pairsToDict()</code></li>
            </ul>
            <h4>Dictionary</h4>
            <ul>
              <li><code>dictLen(), get(key), set(key, value), delete(key)</code></li>
              <li><code>getAtPath(path, notfound), setAtPath(path, value)</code></li>
              <li><code>containsKey(key), mergeInto(dict), walkTree(path)</code></li>
              <li><code>keys(), values(), toPairs()</code></li>
            </ul>

            <AnchorHeading level="h3" id="data-structure-access">Data Structure Access</AnchorHeading>
            <CodeBlock code={`// List access
local numbers = [1, 2, 4]
println(numbers[2]) // 1 based indexing
numbers[1] = 8

// Dictionary access
local animalInfo = { "Animal": "Tiger", "Scientific Name": "Panthera tigris" }
println(animalInfo.get("Scientific Name", "Not found"))`} />

            <AnchorHeading level="h3" id="list-lambdas">List Lambdas</AnchorHeading>
            <p>Inspired by Kotlin, list lambdas allow for list manipulation.</p>

            <h4>Map lambda</h4>
            <CodeBlock code={`local numbers = [1, 2, 3]
// Double all the numbers
local doubled = numbers.map { n -> n * 2 }
println(doubled)  // Output: [2, 4, 6]`} />

            <h4>Filter lambda</h4>
            <CodeBlock code={`local numbers = [1, 2, 3, 4]
// Filter for even numbers
local evens = numbers.filter { n -> n % 2 == 0 }
println(evens)  // Output: [2, 4]`} />

            <h4>Sort lambda</h4>
            <CodeBlock code={`local names = ["Bob", "Alice", "John"]
// Sort names in descending order
local namesSorted = names
  .sort { m, n -> m.textLen() > m.textLen() }
println(namesSorted) // Output:  ["John", "Alice", "Bob"]`} />

            <h4>Min and Max lambdas</h4>
            <CodeBlock code={`local names = ["Bob", "Alice", "John"]
// Find the longest name
local longestName = names
  .max { m, n -> n.textLen() > m.textLen() }
println(longestName)`} />

            <h4>Reduce lambda</h4>
            <CodeBlock code={`local numbers = [1, 2, 3, 4, 5, 6, 7]
// Sum up all the numbers
local numbersSum  = numbers.reduce(0) { x, valueSoFar -> x + valueSoFar }
println(numbersSum) // Output: 28`} />

            <h4>Example</h4>
            <p>Calculating revenue for lemon priced at $2 each, stripping "N/A":</p>
            <CodeBlock code={`global LemonadeSold = [9, 12, "N/A", 15, 18, "N/A", 8]

func GetTotalRevenue() = {
  this.LemonadeSold
    .filter { n -> n ? number }    // Filters for numeric entries
    .map { n -> n * 2 }	    // Multiply by price
    .reduce(0) { x, soFar -> x + soFar }  // Sum up
}`} />
          </div>
        </div>
      </section>
    </div>
  )
}

export default App
