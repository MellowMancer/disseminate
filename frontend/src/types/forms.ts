export type FormDataState = {
  importData: {};
  twitter: { apiKey: string; apiSecret: string; accessToken: string; accessSecret: string; content: string };
  youtube: { apiKey: string; accessToken: string; title: string; description: string; tags: string };
  instagram: { username: string; password: string; caption: string };
  reddit: { clientId: string; clientSecret: string; username: string; password: string; subreddit: string; title: string; content: string };
  mastodon: { instanceUrl: string; accessToken: string; content: string };
  artstation: { username: string; password: string; title: string; description: string };
};

// This type defines the props for our modular tab components
export type TabComponentProps<P extends keyof FormDataState> = {
  data: FormDataState[P];
  handleChange: <F extends keyof FormDataState[P]>(
    platform: P,
    field: F
  ) => (event: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement>) => void;
};