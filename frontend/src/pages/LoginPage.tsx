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
import { Card, CardHeader, CardTitle } from "@/components/ui/card";
import { useLocation, useNavigate } from "react-router-dom";
import { useAuth } from "@/context/AuthContext";

type LoginFormValues = {
  email: string;
  password: string;
};

export default function LoginPage() {
  const { authenticated: _, setAuthenticated } = useAuth();
  const navigate = useNavigate();
  const location = useLocation();
  const from = (location.state as any)?.from?.pathname || "/";

  const form = useForm<LoginFormValues>({
    mode: "onChange",
  });

  const onSubmit: SubmitHandler<LoginFormValues> = async (data) => {
    console.log("Submitting login form with data:", data);
    try {
      const response = await fetch('/auth/login', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          email: data.email,
          password: data.password,
        }),
        credentials: 'include',
      });

      if (response.ok) {
        setAuthenticated(true);
        navigate(from, { replace: true });
      } else {
        const errData = await response.json();
        alert(`Login failed: ${errData.message}`);
      }
    } catch (error) {
      alert("Login error: " + error);
    }
  };

  return (
    <div className="min-h-screen w-screen grid place-items-center bg-background py-16 px-8 md:py-8 md:px-24 mx-auto">
      <Card className="h-min">
        <CardHeader>
          <CardTitle>Log in to Disseminate</CardTitle>
        </CardHeader>
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
            <Button
              type="submit"
              className="w-full py-3 font-medium rounded-md bg-gray-900 text-white hover:bg-gray-800 transition"
            >
              Log In
            </Button>
          </form>
        </Form>
        <p className="mt-2 text-center text-gray-600 text-sm space-y-1">
          <span>
            Don't have an account?
            <button
              onClick={() => navigate("/signup")}
              className="font-semibold text-gray-900 hover:underline focus:outline-none"
            >
              Sign Up
            </button>
          </span>

          <div className="block mt-6 text-gray-700 text-xs italic max-w-xs mx-auto text-start">
            Disemminate is a tool that lets users post on multiple social media from a single dashboard.
          </div>
          <div className="block text-gray-700 text-xs italic max-w-xs mx-auto text-start">
            Supports Bluesky, Twitter, Instagram, Youtube, Mastodon, Artstation, Reddit
          </div>
          <div className="block text-gray-700 text-xs italic max-w-xs mx-auto text-start">
            Users will need to set up their developer accounts to use Disseminate
          </div>
        </p>
      </Card>
    </div>
  );
}
