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

type FormValues = {
  files: FileList;
};

export default function App() {
  const form = useForm<FormValues>({
    mode: "onChange",
  });

  const onSubmit: SubmitHandler<FormValues> = (data) => {
    // TODO: Upload to server
    alert(`${data.files.length} file(s) selected`);
  };

  return (
    <Form {...form}>
      <form
        onSubmit={form.handleSubmit(onSubmit)}
        className="max-w-md mx-auto mt-10 space-y-6 bg-white rounded-xl shadow p-6"
      >
        <FormField
          control={form.control}
          name="files"
          rules={{
            required: "Please select at least one file.",
          }}
          render={({ field }) => (
            <FormItem>
              <FormLabel>Select File(s)</FormLabel>
              <FormControl>
                <Input
                  type="file"
                  multiple
                  onChange={(e: { target: { files: any; }; }) => field.onChange(e.target.files)}
                  className="file:bg-gray-100 file:border-none file:rounded-md file:py-2 file:px-4 file:text-sm file:font-medium cursor-pointer"
                />
              </FormControl>
              <FormDescription>
                You can choose one or more files.
              </FormDescription>
              <FormMessage />
            </FormItem>
          )}
        />
        <Button type="submit">Upload</Button>
        {form.watch("files") &&
          form.watch("files").length > 0 && (
            <div className="mt-4 text-sm">
              <strong>Selected files:</strong>
              <ul className="list-inside list-disc">
                {Array.from(form.watch("files")).map(
                  (file: File, idx) => (
                    <li key={idx}>
                      {file.name} ({file.size} bytes)
                    </li>
                  )
                )}
              </ul>
            </div>
          )}
      </form>
    </Form>
  );
}
