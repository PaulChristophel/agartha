import { json } from '@codemirror/lang-json';
import { yaml } from '@codemirror/lang-yaml';
import CodeMirror from '@uiw/react-codemirror';

interface PillarEditorProps {
  value: string;
  onChange: (value: string) => void;
  onBlur: () => void;
}

export default function PillarEditor({ value, onChange, onBlur }: PillarEditorProps) {
  const language = value.trim().startsWith('{') || value.trim().startsWith('[') ? json() : yaml();

  return <CodeMirror value={value} extensions={[language]} onChange={onChange} onBlur={onBlur} />;
}
