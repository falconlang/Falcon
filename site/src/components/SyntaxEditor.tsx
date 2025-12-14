import { useEffect } from 'react';
import Editor from 'react-simple-code-editor';
import Prism from 'prismjs';
import './SyntaxEditor.css';

interface SyntaxEditorProps {
    value: string;
    onChange: (value: string) => void;
    placeholder?: string;
}

const BRACKET_PAIRS: Record<string, string> = {
    '(': ')',
    '[': ']',
    '{': '}',
    '"': '"',
    "'": "'"
};

export function SyntaxEditor({ value, onChange, placeholder }: SyntaxEditorProps) {
    useEffect(() => {
        // Make sure Prism is initialized
        if (typeof Prism !== 'undefined') {
            Prism.highlightAll();
        }
    }, []);

    const highlight = (code: string) => {
        return Prism.highlight(code, Prism.languages.falcon, 'falcon');
    };

    const handleValueChange = (newValue: string) => {
        onChange(newValue);
    };

    const handleKeyDown = (e: React.KeyboardEvent<HTMLTextAreaElement>) => {
        const target = e.currentTarget;
        const { selectionStart, selectionEnd } = target;
        const currentValue = target.value;

        // Handle Tab key
        if (e.key === 'Tab') {
            e.preventDefault();
            const beforeTab = currentValue.substring(0, selectionStart);
            const afterTab = currentValue.substring(selectionEnd);
            const newValue = beforeTab + '  ' + afterTab; // 2 spaces
            onChange(newValue);

            // Set cursor position after the tab
            setTimeout(() => {
                target.selectionStart = target.selectionEnd = selectionStart + 2;
            }, 0);
            return;
        }

        // Handle auto-closing brackets and quotes
        const char = e.key;
        if (BRACKET_PAIRS[char] && selectionStart === selectionEnd) {
            e.preventDefault();
            const closingChar = BRACKET_PAIRS[char];
            const beforeCursor = currentValue.substring(0, selectionStart);
            const afterCursor = currentValue.substring(selectionEnd);
            const newValue = beforeCursor + char + closingChar + afterCursor;
            onChange(newValue);

            // Set cursor position between the brackets
            setTimeout(() => {
                target.selectionStart = target.selectionEnd = selectionStart + 1;
            }, 0);
            return;
        }

        // Handle Enter key inside brackets - auto-indent
        if (e.key === 'Enter') {
            const beforeCursor = currentValue.substring(0, selectionStart);
            const afterCursor = currentValue.substring(selectionStart);
            const lastChar = beforeCursor[beforeCursor.length - 1];
            const nextChar = afterCursor[0];

            if (lastChar === '{' && nextChar === '}') {
                e.preventDefault();
                const indent = '  ';
                const newValue = beforeCursor + '\n' + indent + '\n' + afterCursor;
                onChange(newValue);

                setTimeout(() => {
                    target.selectionStart = target.selectionEnd = selectionStart + 1 + indent.length;
                }, 0);
                return;
            }
        }

        // Skip closing bracket if we're already at one
        const closingChars = Object.values(BRACKET_PAIRS);
        if (closingChars.includes(char) && currentValue[selectionStart] === char) {
            e.preventDefault();
            target.selectionStart = target.selectionEnd = selectionStart + 1;
            return;
        }
    };

    return (
        <div className="syntax-editor-wrapper">
            <Editor
                value={value}
                onValueChange={handleValueChange}
                highlight={highlight}
                padding={0}
                placeholder={placeholder}
                textareaClassName="syntax-editor-textarea"
                preClassName="syntax-editor-pre"
                onKeyDown={handleKeyDown as any}
                style={{
                    fontFamily: 'IosevkaMonospace, monospace',
                    fontSize: '1rem',
                    lineHeight: '1.5',
                    height: '100%',
                    overflow: 'auto'
                }}
            />
        </div>
    );
}
