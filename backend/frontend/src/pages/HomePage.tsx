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
                <h1 className="text-3xl md:text-5xl font-semibold mb-2 md:mb-6 text-highlight">Disseminate</h1>
                <p className="text-l md:text-xl text-secondary-text mb-8">
                    Post to multiple social media platforms from a single dashboard.
                    Supports Bluesky, Twitter, Instagram, Youtube, Mastodon, Artstation, Reddit
                </p>
                <nav className="space-x-4">
                    <Button
                        onClick={handleStartPostingClick}
                    >
                        Start Posting
                    </Button>
                </nav>
            </div>

            {/* Right half */}
            <div className="md:w-1/2 flex flex-col justify-center text-gray-900">
                <DynamicShadowWrapper>
                    <Card>
                        <CardHeader>
                            <CardTitle>Upload your content</CardTitle>
                        </CardHeader>
                        <Form {...form}>
                            <form
                                onSubmit={form.handleSubmit(onSubmit)}
                                className="max-w-lg mx-auto space-y-6"
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
                                            <FormItem>
                                                <FormLabel className="text-highlight font-medium">Select File(s)</FormLabel>
                                                <FormControl className="flex-col justify-center border-1 m-0 p-0 h-10">
                                                    <Input
                                                        type="file"
                                                        multiple
                                                        onChange={handleFileChange}
                                                        className="file:bg-button file:text-button-text file:border-none file:rounded-md file:h-full file:py-2 file:px-2 file:mr-2 file:text-sm file:font-medium cursor-pointer flex-col justify-center text-secondary-text border-card-outline"
                                                        accept=".mp4, .avi, .mkv, .ogg, .webm, .m4v, .mov, .jpg, .webp, .HEIC, .HEIC, .png, .jpeg, .tiff "
                                                    />
                                                </FormControl>
                                                <FormDescription className="text-secondary-text">You can choose one or more files. <div className="text-s text-secondary-text"><i>Supports .mp4, .avi, .mkv, .ogg, .webm, .m4v, .mov, .jpg, .webp, .HEIC, .HEIC, .png, .jpeg, .tiff </i></div></FormDescription>
                                                <FormMessage />
                                            </FormItem>
                                        )
                                    }}
                                />

                                {form.watch("files") && form.watch("files").length > 0 && (
                                    <div className="mt-4 text-sm text-secondary-text">
                                        <strong>Selected files:</strong>
                                        <ul className="list-inside list-disc">
                                            {Array.from(form.watch("files")).map((file) => (
                                                <li key={file.name}>
                                                    {file.name} ({(file.size / 1024 / 1024).toFixed(2)} MB)
                                                </li>
                                            ))}
                                        </ul>
                                    </div>
                                )}

                                <Button type="submit" className="w-max bg-button text-button-text">
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