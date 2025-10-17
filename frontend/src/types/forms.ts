export type FormDataState = {
  importData: {};
  twitter: { content: string };
  youtube: { description: string; tags: string };
  instagram: { caption: string };
  reddit: {  };
  mastodon: {  };
  artstation: {  };
};

// This type defines the props for our modular tab components
export type TabComponentProps<P extends keyof FormDataState> = {
  data: FormDataState[P];
  handleChange: <F extends keyof FormDataState[P]>(
    platform: P,
    field: F
  ) => (event: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement>) => void;
};