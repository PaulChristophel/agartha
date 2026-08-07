import yaml from 'react-syntax-highlighter/dist/esm/languages/prism/yaml';
import { PrismLight as SyntaxHighlighter } from 'react-syntax-highlighter';
import { solarizedDarkAtom } from 'react-syntax-highlighter/dist/esm/styles/prism';

SyntaxHighlighter.registerLanguage('yaml', yaml);

export default function OutputRenderer({ output }: { output: string }) {
  return (
    <SyntaxHighlighter language="yaml" style={solarizedDarkAtom}>
      {output}
    </SyntaxHighlighter>
  );
}
