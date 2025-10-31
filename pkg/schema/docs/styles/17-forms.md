## 17. 📝 Forms & Validation

### 17.1 Form Design Principles

1. **Clear labels** — Every field needs a label
2. **Smart defaults** — Pre-fill when possible
3. **Inline validation** — Provide immediate feedback
4. **Error prevention** — Guide users to success
5. **Progressive disclosure** — Show complexity gradually

### 17.2 Form Field Components

```typescript
// Text input
<FormField>
  <FormLabel htmlFor="name">Name</FormLabel>
  <FormControl>
    <Input 
      id="name"
      placeholder="John Doe"
      required
    />
  </FormControl>
  <FormDescription>Your full legal name</FormDescription>
  <FormMessage />
</FormField>

// Select
<FormField>
  <FormLabel>Country</FormLabel>
  <FormControl>
    <Select>
      <SelectTrigger>
        <SelectValue placeholder="Select country" />
      </SelectTrigger>
      <SelectContent>
        <SelectItem value="us">United States</SelectItem>
        <SelectItem value="uk">United Kingdom</SelectItem>
        <SelectItem value="ca">Canada</SelectItem>
      </SelectContent>
    </Select>
  </FormControl>
</FormField>

// Checkbox
<FormField>
  <div className="flex items-center gap-2">
    <FormControl>
      <Checkbox id="terms" />
    </FormControl>
    <FormLabel htmlFor="terms">
      I agree to the terms and conditions
    </FormLabel>
  </div>
</FormField>

// Radio group
<FormField>
  <FormLabel>Notification preference</FormLabel>
  <FormControl>
    <RadioGroup defaultValue="email">
      <div className="flex items-center gap-2">
        <RadioGroupItem value="email" id="email" />
        <Label htmlFor="email">Email</Label>
      </div>
      <div className="flex items-center gap-2">
        <RadioGroupItem value="sms" id="sms" />
        <Label htmlFor="sms">SMS</Label>
      </div>
    </RadioGroup>
  </FormControl>
</FormField>
```

### 17.3 Validation Patterns

```typescript
// Validation rules
interface ValidationRule {
  required?: boolean | string;
  minLength?: { value: number; message: string };
  maxLength?: { value: number; message: string };
  pattern?: { value: RegExp; message: string };
  validate?: (value: any) => boolean | string;
}

// Common validators
const validators = {
  required: (message = 'This field is required') => ({
    required: message,
  }),
  
  email: () => ({
    pattern: {
      value: /^[A-Z0-9._%+-]+@[A-Z0-9.-]+\.[A-Z]{2,}$/i,
      message: 'Invalid email address',
    },
  }),
  
  minLength: (length: number) => ({
    minLength: {
      value: length,
      message: `Must be at least ${length} characters`,
    },
  }),
  
  maxLength: (length: number) => ({
    maxLength: {
      value: length,
      message: `Must be no more than ${length} characters`,
    },
  }),
  
  password: () => ({
    minLength: { value: 8, message: 'Password must be at least 8 characters' },
    pattern: {
      value: /^(?=.*[a-z])(?=.*[A-Z])(?=.*\d)/,
      message: 'Password must contain uppercase, lowercase, and number',
    },
  }),
};

// Usage
<FormField
  name="email"
  rules={{
    ...validators.required(),
    ...validators.email(),
  }}
>
  <Input type="email" />
</FormField>
```

### 17.4 Form Submission

```typescript
function ContactForm() {
  const [isSubmitting, setIsSubmitting] = React.useState(false);
  const [submitError, setSubmitError] = React.useState<string | null>(null);
  const [submitSuccess, setSubmitSuccess] = React.useState(false);
  
  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setIsSubmitting(true);
    setSubmitError(null);
    
    const formData = new FormData(e.target as HTMLFormElement);
    const data = Object.fromEntries(formData);
    
    try {
      await submitContactForm(data);
      setSubmitSuccess(true);
      (e.target as HTMLFormElement).reset();
    } catch (error) {
      setSubmitError('Failed to submit form. Please try again.');
    } finally {
      setIsSubmitting(false);
    }
  };
  
  return (
    <form onSubmit={handleSubmit}>
      {submitSuccess && (
        <Alert variant="success">
          <AlertTitle>Success!</AlertTitle>
          <AlertDescription>
            Your message has been sent.
          </AlertDescription>
        </Alert>
      )}
      
      {submitError && (
        <Alert variant="destructive">
          <AlertTitle>Error</AlertTitle>
          <AlertDescription>{submitError}</AlertDescription>
        </Alert>
      )}
      
      <FormField>
        <FormLabel>Email</FormLabel>
        <Input name="email" type="email" required />
      </FormField>
      
      <FormField>
        <FormLabel>Message</FormLabel>
        <Textarea name="message" required />
      </FormField>
      
      <Button type="submit" disabled={isSubmitting}>
        {isSubmitting ? (
          <>
            <Icon name="loader" className="icon-spin" />
            Sending...
          </>
        ) : (
          'Send Message'
        )}
      </Button>
    </form>
  );
}
```

### 17.5 Multi-Step Forms

```typescript
function MultiStepForm() {
  const [step, setStep] = React.useState(1);
  const [formData, setFormData] = React.useState({});
  
  const updateFormData = (data: Record<string, any>) => {
    setFormData(prev => ({ ...prev, ...data }));
  };
  
  const nextStep = () => setStep(prev => prev + 1);
  const prevStep = () => setStep(prev => prev - 1);
  
  return (
    <div>
      <StepIndicator currentStep={step} totalSteps={3} />
      
      {step === 1 && (
        <Step1 data={formData} onNext={(data) => {
          updateFormData(data);
          nextStep();
        }} />
      )}
      
      {step === 2 && (
        <Step2 
          data={formData} 
          onNext={(data) => {
            updateFormData(data);
            nextStep();
          }}
          onBack={prevStep}
        />
      )}
      
      {step === 3 && (
        <Step3 
          data={formData}
          onSubmit={async (data) => {
            updateFormData(data);
            await submitForm({ ...formData, ...data });
          }}
          onBack={prevStep}
        />
      )}
    </div>
  );
}
```

---

