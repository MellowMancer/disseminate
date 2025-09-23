import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Textarea } from '@/components/ui/textarea';
import type { TabComponentProps } from '@/types/forms';

// This component can be used for both Video and Short tabs
export function YouTubeTab({ data, handleChange, platform }: TabComponentProps<'youtubeVideo' | 'youtubeShort'> & { platform: 'youtubeVideo' | 'youtubeShort' }) {
  return (
    <div className="mt-4 space-y-4">
      <div><Label>Title</Label><Input value={data.title} onChange={handleChange(platform, 'title')} /></div>
      <div><Label>Description</Label><Textarea value={data.description} onChange={handleChange(platform, 'description')} /></div>
      <div><Label>Client ID</Label><Input value={data.clientId} onChange={handleChange(platform, 'clientId')} /></div>
      <div><Label>Client Secret</Label><Input value={data.clientSecret} onChange={handleChange(platform, 'clientSecret')} /></div>
    </div>
  );
}