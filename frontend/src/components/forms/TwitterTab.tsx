import React from 'react';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Textarea } from '@/components/ui/textarea';
import type { FormDataState, TabComponentProps } from '@/types/forms';
import { InputWithLabel } from '../ui/inputWithLabel';

export function TwitterTab({ data, handleChange }: TabComponentProps<'twitter'>) {
  return (
    <div className="mt-4 space-y-4">
      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
        <InputWithLabel type='text' label='API Key' value={data.apiKey} onChange={handleChange('twitter', 'apiKey')} />
        <InputWithLabel type='text' label='API Secret Key' value={data.apiSecret} onChange={handleChange('twitter', 'apiSecret')} />
        <InputWithLabel type='text' label='API Secret Key' value={data.accessToken} onChange={handleChange('twitter', 'accessToken')} />
        <InputWithLabel type='text' label='API Secret Key' value={data.accessSecret} onChange={handleChange('twitter', 'accessSecret')} />
        </div>
      <div><Label>Tweet Content</Label><Textarea value={data.content} onChange={handleChange('twitter', 'content')} /></div>
    </div>
  );
}