import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import * as React from "react"

export interface InputWithLabelProps extends React.ComponentProps<"input"> {
  label: string;
}

export function InputWithLabel({ className, type, label, placeholder, ...props }: InputWithLabelProps) {
  const inputId = label; 

  return (
    <div className="grid w-full max-w-sm items-center gap-3">
      <Label htmlFor={inputId}>{label}</Label>
      <Input type={type} id={inputId} placeholder={placeholder} {...props}/>
    </div>
  )
}