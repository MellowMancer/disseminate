import { useForm, type SubmitHandler } from "react-hook-form";
import {
    Form,
    FormItem,
    FormLabel,
    FormControl,
    FormDescription,
    FormMessage,
    FormField,
} from "@/components/ui/form";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import { Card, CardHeader, CardTitle } from '@/components/ui/card';
import { toast } from 'sonner';
import { useNavigate } from 'react-router-dom';
import { DynamicShadowWrapper } from "@/components/ui/dynamic-shadow-wrapper";

type FormValues = {
    files: FileList;
};

export default function HomePage() {
    const navigate = useNavigate();

    const form = useForm<FormValues>({
        mode: "onChange",
    });

    const onSubmit: SubmitHandler<FormValues> = (data) => {
        navigate('/schedule', { state: { files: data.files } });
    };

    const handleStartPostingClick = () => {
        toast.success('Yeah, the button does not do much, I just kept it because it makes the UI look prettier');
    };

    return (
        <div className="flex flex-col md:flex-row w-full h-full">
            {/* Left half */}
            <div className="md:w-1/2 flex flex-col justify-center pr-8 md:pr-16 mb-16 md:mb-0">
                <h1 className="text-4xl md:text-6xl font-bold mb-4 md:mb-6 text-primary">Disseminate</h1>
                <p className="text-base md:text-lg text-muted-foreground mb-8 leading-relaxed">
                    Post to multiple social media platforms from a single dashboard.
                    Supports Bluesky, Twitter, Instagram, Youtube, Mastodon, Artstation, Reddit
                </p>
                <nav>
                    <Button
                        onClick={handleStartPostingClick}
                        size="lg"
                    >
                        Start Posting
                    </Button>
                </nav>
            </div>

            {/* Right half */}
            <div className="md:w-1/2 flex flex-col justify-center text-gray-900">
                <DynamicShadowWrapper>
                    <Card>
                        <CardHeader className="space-y-2">
                            <CardTitle className="text-2xl">Upload your content</CardTitle>
                        </CardHeader>
                        <Form {...form}>
                            <form
                                onSubmit={form.handleSubmit(onSubmit)}
                                className="max-w-lg mx-auto space-y-6 px-6 pb-6"
                            >
                                <FormField
                                    control={form.control}
                                    name="files"
                                    rules={{
                                        required: "Please select at least one file.",
                                        validate: files => files.length <= 13 || "You can upload a maximum of 10 files."
                                    }}
                                    render={({ field }) => {
                                        const handleFileChange = (e: { target: { files: any; value: string; }; }) => {
                                            const files = e.target.files;
                                            if (files.length > 12) {
                                                e.target.value = "";
                                                toast.error('A maximum of 12 inputs are supported');
                                                return;
                                            }

                                            field.onChange(files);
                                        };
                                        return (
                                            <FormItem className="space-y-3">
                                                <FormLabel className="text-sm font-medium">Select File(s)</FormLabel>
                                                <FormControl className="flex-col justify-center border-1 m-0 p-0 h-10">
                                                    <Input
                                                        type="file"
                                                        multiple
                                                        onChange={handleFileChange}
                                                        className="file:bg-primary file:text-primary-foreground file:border-none file:rounded-md file:h-full file:py-2 file:px-3 file:mr-3 file:text-sm file:font-medium cursor-pointer flex-col justify-center text-sm text-muted-foreground border-border"
                                                        accept=".mp4, .avi, .mkv, .ogg, .webm, .m4v, .mov, .jpg, .webp, .HEIC, .HEIC, .png, .jpeg, .tiff "
                                                    />
                                                </FormControl>
                                                <FormDescription className="text-sm text-muted-foreground leading-relaxed">
                                                    You can choose one or more files.
                                                    <div className="text-xs text-muted-foreground mt-1"><i>Supports .mp4, .avi, .mkv, .ogg, .webm, .m4v, .mov, .jpg, .webp, .HEIC, .png, .jpeg, .tiff</i></div>
                                                </FormDescription>
                                                <FormMessage className="text-sm" />
                                            </FormItem>
                                        )
                                    }}
                                />

                                {form.watch("files") && form.watch("files").length > 0 && (
                                    <div className="space-y-2">
                                        <p className="text-sm font-medium text-primary">Selected files:</p>
                                        <ul className="list-inside list-disc space-y-1 text-sm text-muted-foreground">
                                            {Array.from(form.watch("files")).map((file) => (
                                                <li key={file.name}>
                                                    {file.name} ({(file.size / 1024 / 1024).toFixed(2)} MB)
                                                </li>
                                            ))}
                                        </ul>
                                    </div>
                                )}

                                <Button type="submit" size="lg" className="w-full bg-primary text-primary-foreground">
                                    Upload
                                </Button>
                            </form>
                        </Form>
                    </Card>
                </DynamicShadowWrapper>
            </div>
        </div>
    );
}