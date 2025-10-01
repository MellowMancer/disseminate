import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Textarea } from '@/components/ui/textarea';
import type { TabComponentProps } from '@/types/forms';

// This component can be used for both Video and Short tabs
export function YouTubeTab({ data, handleChange }: Readonly<TabComponentProps<'youtube'>>) {
  return (
    <div className="mt-4 space-y-4">
      <div><Label>Title</Label><Input value={data.title} onChange={handleChange('youtube', 'title')} /></div>
      <div><Label>Description</Label><Textarea value={data.description} onChange={handleChange('youtube', 'description')} /></div>
    </div>
  );
}