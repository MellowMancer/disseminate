import { Label } from '@/components/ui/label';
import { Textarea } from '@/components/ui/textarea';
import type { TabComponentProps } from '@/types/forms';

export function InstagramTab({ data, handleChange }: Readonly<TabComponentProps<'instagram'>>) {
  return (
    <div className="mt-4 space-y-4">
      <div><Label>Caption</Label><Textarea value={data.caption} onChange={handleChange('instagram', 'caption')} /></div>
    </div>
  );
}