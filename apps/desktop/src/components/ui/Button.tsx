import React, { ButtonHTMLAttributes } from "react";
import { cn } from "../../utils/cn";
import { cva, VariantProps } from "class-variance-authority";

const ButtonStyles = cva(
  [
    "font-semibold",
    "outline-sky-400",
    "rounded-md",
    "w-fit",
    "flex",
    " items-center",
    "transition-colors",
  ],
  {
    variants: {
      intent: {
        primary: [
          "bg-[var(--color-primary)]",
          "text-white",
          "border-transparent",
          "hover:bg-[var(--color-primary)]/70",
          "font-normal"
        ],
        // **or**
        // primary: "bg-blue-500 text-white border-transparent hover:bg-blue-600",
        secondary: [
          // "bg-[var(--color-secondary)]",
          "text-white",
          // "border-gray-400",
          "hover:bg-[var(--color-secondary)]/70",
          "font-normal"
        ],
        ghost: [
          "bg-transparent",
          "text-gray-900",
          "border-transparent",
          "hover:bg-white",
        ],
        destructive: [
          "bg-[var(--color-destructive)]",
          "text-white",
          "border-transparent",
          "hover:bg-[var(--color-destructive)]/80",
          "hover:text-white",
        ],
        cta: [
          "bg-gradient-to-r from-green-200 to-blue-500",
          "text-white",
          "border-transparent",
        ],
        action: [
          "bg-sky-400",
          "text-white",
          "border-transparent",
          "w-full",
          "justify-center",
        ],
        window: [
          "bg-transparent",
          "text-white",
          "border-transparent",
          "w-full",
          "justify-center",
        ],
      },
      size: {
        small: ["text-xs", "py-0.5", "px-1"],
        md: ["text-sm", "py-1", "px-3"],
        sm: "h-9 rounded-md px-3",
        medium: ["text-base", "py-2", "px-4"],
        icon: ["p-2", "drop-shadow", 'rounded-full'],
        cta: [
          "text-sm",
          "py-1",
          "px-3",
          "rounded-md",
          "w-full",
        ],
        menu: [
          "text-sm",
          "py-2",
          "px-4",
          "w-full",
          "rounded-full",
          "gap-3"
        ],
        window: [
          "p-2",
          "drop-shadow",
          "rounded-full",
        ],
      },
    },
    // compoundVariants: [
    //     {
    //         intent: "ghost",
    //         class: "uppercase",
    //         size: "small",
    //         // **or** if you're a React.js user, `className` may feel more consistent:
    //         // className: "uppercase"
    //     },
    // ],
    defaultVariants: {
      intent: "ghost",
      size: "icon",
    },
  },
);

type ButtonProps = ButtonHTMLAttributes<HTMLButtonElement> &
  VariantProps<typeof ButtonStyles> & {
    className?: string;
    children?: React.ReactNode;
  };

const Button = ({ children, className, intent, size, ...props }: ButtonProps) => {
  return (
    <button
      className={cn(
        ButtonStyles({ intent, size, className }),
      )}
      type="button"
      {...props}
    >
      {children}
    </button>
  );
};

export default Button;
