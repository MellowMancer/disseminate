import { Label } from '@/components/ui/label';
import { Textarea } from '@/components/ui/textarea';
import type { TabComponentProps } from '@/types/forms';

export function TwitterTab({ data, handleChange }: Readonly<TabComponentProps<'twitter'>>) {
  return (
    <div className="mt-4 space-y-4">
      <div><Label>Tweet Content</Label><Textarea value={data.content} onChange={handleChange('twitter', 'content')} /></div>
    </div>
  );
}