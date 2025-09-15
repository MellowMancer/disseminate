import React, { useState } from "react";
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

type SignUpFormValues = {
  email: string;
  password: string;
  confirmPassword?: string;
};

type LoginFormValues = {
  email: string;
  password: string;
};

export default function CloudAuthPage() {
  const [mode, setMode] = useState<"login" | "signup">("login");
  const [emailVerified, setEmailVerified] = useState(false);
  const form = useForm<SignUpFormValues | LoginFormValues>({
    mode: "onChange",
  });

  const onSubmit: SubmitHandler<SignUpFormValues | LoginFormValues> = (data) => {
    if (mode === "signup") {
      alert("Verification email sent. Please check your inbox.");
      setEmailVerified(false);
    } else {
      setEmailVerified(true);
    }
  };

  const switchMode = () => {
    form.reset();
    setEmailVerified(false);
    setMode((prev) => (prev === "login" ? "signup" : "login"));
  };

  return (
    <div className="min-h-screen flex flex-col items-center justify-center bg-gray-100 p-6">
      <div className="w-full max-w-md bg-white rounded-lg shadow-md p-8 sm:p-10 ring-1 ring-gray-300">
        <h2 className="text-2xl font-semibold text-gray-900 mb-6 text-center">
          {mode === "login" ? "Log in to Disseminate" : "Create a Disseminate account"}
        </h2>

        {!emailVerified ? (
          <Form {...form}>
            <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-6" noValidate>
              <FormField
                control={form.control}
                name="email"
                rules={{
                  required: "Please enter your email",
                  pattern: {
                    value: /^\S+@\S+$/i,
                    message: "Invalid email address",
                  },
                }}
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Email address</FormLabel>
                    <FormControl>
                      <Input type="email" placeholder="you@example.com" {...field} />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />

              <FormField
                control={form.control}
                name="password"
                rules={{ required: "Please enter your password", minLength: 6 }}
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Password</FormLabel>
                    <FormControl>
                      <Input type="password" placeholder="••••••••" {...field} />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />

              {mode === "signup" && (
                <FormField
                  control={form.control}
                  name="confirmPassword"
                  rules={{
                    required: "Please confirm your password",
                    validate: (value) =>
                      value === form.getValues("password") || "Passwords do not match",
                  }}
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel>Confirm password</FormLabel>
                      <FormControl>
                        <Input type="password" placeholder="••••••••" {...field} />
                      </FormControl>
                      <FormMessage />
                    </FormItem>
                  )}
                />
              )}

              <Button
                type="submit"
                className="w-full py-3 font-medium rounded-md bg-gray-900 text-white hover:bg-gray-800 transition"
              >
                {mode === "login" ? "Log In" : "Sign Up"}
              </Button>
            </form>
          </Form>
        ) : (
          <div className="text-center space-y-4 text-gray-800">
            <p>
              A verification email has been sent to your email address. Please check your inbox and verify your email to continue.
            </p>
            <Button
              onClick={() => setEmailVerified(false)}
              className="px-6 py-3 bg-gray-900 text-white rounded-md hover:bg-gray-800 transition"
            >
              Resend Verification Email
            </Button>
          </div>
        )}

        <p className="mt-8 text-center text-gray-600 text-sm space-y-1">
          <span>
            {mode === "login" ? "Don't have an account?" : "Already have an account?"}{" "}
            <button
              onClick={switchMode}
              className="font-semibold text-gray-900 hover:underline focus:outline-none"
            >
              {mode === "login" ? "Sign Up " : "Log In "}
            </button>
          </span>

          <div>
            Don't wish to create an account?{" "}
            <Link
              to="/local"
              className="font-semibold text-gray-900 hover:underline focus:outline-none"
            >
              Local Account
            </Link>
          </div>

          <span className="block mt-4 text-gray-700 text-xs italic max-w-xs mx-auto">
            Making an account is not necessary if you primarily use a single device. An account just makes the tool convenient to use across multiple devices.
          </span>
        </p>
      </div>
    </div>
  );
}
