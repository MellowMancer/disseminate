import React from "react";
import { useForm, type SubmitHandler } from "react-hook-form";
import {
    Form,
    FormItem,
    FormLabel,
    FormControl,
    FormMessage,
    FormField,
} from "@/components/ui/form";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import { Link } from "react-router-dom";

type LocalSignUpValues = {
    username: string;
};

export default function LocalAccountPage() {
    const form = useForm<LocalSignUpValues>({
        mode: "onChange",
    });

    const onSubmit: SubmitHandler<LocalSignUpValues> = (data) => {
        alert(`Local account created with username: ${data.username}`);
        // TODO: Local account creation logic
    };

    return (
        <div className="min-h-screen flex items-center justify-center bg-gray-100 p-6">
            <div className="w-full max-w-md bg-white rounded-lg shadow-md p-8 sm:p-10 ring-1 ring-gray-300">
                <h2 className="text-2xl font-semibold text-gray-900 mb-6 text-center">
                    Create a Local Account
                </h2>

                <p className="mb-6 text-sm text-gray-700">
                    <strong>Disclaimer:</strong> Local accounts store data only on the current device. Switching devices or clearing browser data means you’ll need to re-enter developer keys and settings. No cloud sync.
                </p>


                <Form {...form}>
                    <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-6" noValidate>
                        <FormField
                            control={form.control}
                            name="username"
                            rules={{
                                required: "Username is required",
                                minLength: { value: 3, message: "Minimum 3 characters" },
                            }}
                            render={({ field }) => (
                                <FormItem>
                                    <FormLabel>Username</FormLabel>
                                    <FormControl>
                                        <Input type="text" placeholder="Your username" {...field} />
                                    </FormControl>
                                    <FormMessage />
                                </FormItem>
                            )}
                        />

                        <Button
                            type="submit"
                            className="w-full py-3 font-medium rounded-md bg-gray-900 text-white hover:bg-gray-800 transition"
                        >
                            Create/Login to Local Account
                        </Button>
                    </form>
                </Form>

                <div className="mt-8 space-y-3 text-center text-gray-600 text-sm">
                    <p>
                        Wish to{" "}
                        <Link to="/auth/" className="font-semibold text-gray-900 hover:underline">
                            log in/sign up
                        </Link>{" "}
                        instead?
                    </p>
                </div>
            </div>
        </div>
    );
}
