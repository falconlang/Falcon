import { useEffect, useState } from 'react';
import Prism from 'prismjs';
import './CodeBlock.css';

// Register Falcon language
Prism.languages.falcon = {
    'comment': /\/\/.*$/m,
    'string': {
        pattern: /"(?:[^\\"]|\\.)*"/,
        greedy: true
    },
    'keyword': /\b(?:true|false|if|else|for|step|in|while|do|break|walkAll|global|local|compute|this|func|when|any|undefined)\b/,
    'boolean': /\b(?:true|false)\b/,
    'number': /\b\d+(?:\.\d+)?(?:[eE][+-]?\d+)?\b|0x[\da-fA-F]+|0b[01]+/,
    'operator': /::|\.\.|->|===|!==|==|!=|<=|>=|<<|>>|&&|\|\||[+\-*\/%^&|~<>=_@:!?]/,
    'punctuation': /[{}[\]();,.]/,
    'builtin': /\b(?:println|openScreen|openScreenWithValue|closeScreenWithValue|getStartValue|closeScreen|closeApp|getPlainStartText|dec|bin|octal|hexa|sqrt|abs|neg|log|exp|round|ceil|floor|sin|cos|tan|asin|acos|atan|atan2|degrees|radians|decToHex|decToBin|hexToDec|binToDec|randInt|randFloat|setRandSeed|min|max|avgOf|maxOf|minOf|geoMeanOf|stdDevOf|stdErrOf|modeOf|mod|rem|quot|formatDecimal|copyList|copyDict|makeColor|splitColor|textLen|trim|uppercase|lowercase|startsWith|contains|containsAny|containsAll|split|splitAtFirst|splitAtAny|splitAtFirstOfAny|splitAtSpaces|reverse|csvRowToList|csvTableToList|segment|replace|replaceFrom|replaceFromLongestFirst|listLen|add|containsItem|indexOf|insert|remove|appendList|lookupInPairs|join|slice|random|reverseList|sort|allButFirst|allButLast|pairsToDict|toCsvRow|toCsvTable|dictLen|get|set|delete|getAtPath|setAtPath|containsKey|mergeInto|walkTree|keys|values|toPairs|map|filter|reduce)\b/
};

interface CodeBlockProps {
    code: string;
    language?: string;
}

export function CodeBlock({ code, language = 'falcon' }: CodeBlockProps) {
    const [copied, setCopied] = useState(false);

    useEffect(() => {
        Prism.highlightAll();
    }, [code, language]);

    const handleCopy = async () => {
        try {
            await navigator.clipboard.writeText(code.trim());
            setCopied(true);
            setTimeout(() => setCopied(false), 2000);
        } catch (err) {
            console.error('Failed to copy:', err);
        }
    };

    return (
        <div className="code-block-wrapper">
            <button
                className="code-copy-button"
                onClick={handleCopy}
                title="Copy code"
            >
                {copied ? (
                    <svg width="16" height="16" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                        <path d="M20 6L9 17L4 12" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" />
                    </svg>
                ) : (
                    <svg width="16" height="16" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                        <rect x="9" y="9" width="13" height="13" rx="2" ry="2" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" />
                        <path d="M5 15H4C2.89543 15 2 14.1046 2 13V4C2 2.89543 2.89543 2 4 2H13C14.1046 2 15 2.89543 15 4V5" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" />
                    </svg>
                )}
            </button>
            <pre className={`language-${language}`}>
                <code className={`language-${language}`}>
                    {code.trim()}
                </code>
            </pre>
        </div>
    );
}
