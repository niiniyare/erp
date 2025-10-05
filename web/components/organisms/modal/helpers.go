package modal

import (
	"fmt"

	"github.com/niiniyare/erp/web/components/atoms"
)

// getModalSizeClasses returns Tailwind classes for modal sizing
func getModalSizeClasses(size ModalSize) string {
	switch size {
	case ModalSizeSM:
		return "max-w-sm w-full"
	case ModalSizeMD:
		return "max-w-md w-full"
	case ModalSizeLG:
		return "max-w-lg w-full"
	case ModalSizeXL:
		return "max-w-xl w-full"
	case ModalSize2XL:
		return "max-w-2xl w-full"
	case ModalSizeFull:
		return "w-full h-full max-w-none max-h-none"
	default:
		return "max-w-md w-full" // Default to medium
	}
}

// getModalPositionClasses returns classes for modal positioning
func getModalPositionClasses(position string) string {
	switch position {
	case "top":
		return "items-start pt-16"
	case "bottom":
		return "items-end pb-16"
	case "center", "":
		return "items-center"
	default:
		return "items-center"
	}
}

// getModalBackdropClasses returns backdrop styling classes
func getModalBackdropClasses(backdrop string) string {
	switch backdrop {
	case "blur":
		return "bg-gray-900/50 backdrop-blur-sm"
	case "dark":
		return "bg-gray-900/80"
	case "light":
		return "bg-gray-900/20"
	case "":
		return "bg-gray-900/50"
	default:
		return backdrop // Custom backdrop classes
	}
}

// getModalZIndex returns z-index classes
func getModalZIndex(zIndex string) string {
	if zIndex != "" {
		return zIndex
	}
	return "z-50" // Default modal z-index
}

// getModalAnimation returns animation classes
func getModalAnimation(animation string) string {
	switch animation {
	case "fade":
		return "transition-opacity duration-300"
	case "scale":
		return "transition-all duration-300"
	case "slide":
		return "transition-transform duration-300"
	case "":
		return "transition-all duration-300"
	default:
		return animation // Custom animation classes
	}
}

// getFooterAlignmentClasses returns footer button alignment classes
func getFooterAlignmentClasses(alignment string) string {
	switch alignment {
	case "left":
		return "justify-start"
	case "center":
		return "justify-center"
	case "right":
		return "justify-end"
	case "between":
		return "justify-between"
	case "":
		return "justify-end" // Default to right alignment
	default:
		return "justify-end"
	}
}

// getModalTypeClasses returns type-specific classes
func getModalTypeClasses(modalType ModalType) string {
	switch modalType {
	case ModalTypeConfirm:
		return "text-center"
	case ModalTypeAlert:
		return "text-center"
	case ModalTypeForm, ModalTypeDefault:
		return ""
	default:
		return ""
	}
}

// getConfirmVariantClasses returns styling for confirmation modals
func getConfirmVariantClasses(variant string) (iconColor, buttonVariant string) {
	switch variant {
	case "danger":
		return "text-red-600 dark:text-red-400", "danger"
	case "warning":
		return "text-yellow-600 dark:text-yellow-400", "warning"
	case "info":
		return "text-blue-600 dark:text-blue-400", "primary"
	default:
		return "text-gray-600 dark:text-gray-400", "primary"
	}
}

// getConfirmIcon returns appropriate icon for confirmation variant
func getConfirmIcon(variant string) string {
	switch variant {
	case "danger":
		return "exclamation-triangle"
	case "warning":
		return "exclamation-circle"
	case "info":
		return "information-circle"
	default:
		return "question-mark-circle"
	}
}

// getModalRole returns appropriate ARIA role
func getModalRole(modalType ModalType) string {
	switch modalType {
	case ModalTypeConfirm, ModalTypeAlert:
		return "alertdialog"
	default:
		return "dialog"
	}
}

// generateModalAlpineData creates Alpine.js data object
func generateModalAlpineData(props ModalProps) string {
	openState := "false"
	if props.Open {
		openState = "true"
	}

	return fmt.Sprintf(`{
		open: %s,
		loading: false,
		
		openModal() {
			this.open = true;
			this.$nextTick(() => {
				this.focusFirstElement();
			});
			%s
		},
		
		closeModal() {
			this.open = false;
			%s
		},
		
		handleBackdropClick() {
			if (%t) {
				this.closeModal();
			}
		},
		
		handleEscapeKey() {
			if (%t) {
				this.closeModal();
			}
		},
		
		focusFirstElement() {
			if (%t) {
				const focusable = this.$el.querySelectorAll('button, [href], input, select, textarea, [tabindex]:not([tabindex="-1"])');
				if (focusable.length > 0) {
					focusable[0].focus();
				}
			}
		},
		
		setLoading(loading) {
			this.loading = loading;
		}
	}`,
		openState,
		props.OnOpen,
		props.OnClose,
		props.CloseOnBackdrop,
		props.CloseOnEscape,
		props.FocusTrap,
	)
}

// generateFormAlpineData creates Alpine.js data for form modals
func generateFormAlpineData(props FormModalProps) string {
	return fmt.Sprintf(`{
		open: false,
		loading: false,
		errors: {},
		
		openModal() {
			this.open = true;
			this.errors = {};
			this.$nextTick(() => {
				const firstInput = this.$el.querySelector('input, select, textarea');
				if (firstInput) {
					firstInput.focus();
				}
			});
		},
		
		closeModal() {
			this.open = false;
			this.errors = {};
		},
		
		submitForm() {
			this.loading = true;
			this.errors = {};
		},
		
		handleFormResponse(response) {
			this.loading = false;
			if (response.errors) {
				this.errors = response.errors;
			} else {
				this.closeModal();
			}
		}
	}`)
}

// getButtonVariantFromString converts string to ButtonVariant
func getButtonVariantFromString(variant string) atoms.ButtonVariant {
	switch variant {
	case "primary":
		return atoms.ButtonPrimary
	case "secondary":
		return atoms.ButtonSecondary
	case "success":
		return atoms.ButtonSuccess
	case "danger":
		return atoms.ButtonDanger
	case "warning":
		return atoms.ButtonWarning
	case "info":
		return atoms.ButtonInfo
	case "light":
		return atoms.ButtonLight
	case "dark":
		return atoms.ButtonDark
	case "ghost":
		return atoms.ButtonGhost
	default:
		return atoms.ButtonSecondary
	}
}

// getDefaultModalProps returns sensible defaults for modal props
func getDefaultModalProps() ModalProps {
	return ModalProps{
		Size:            ModalSizeMD,
		Type:            ModalTypeDefault,
		Closable:        true,
		CloseOnBackdrop: true,
		CloseOnEscape:   true,
		FocusTrap:       true,
		Position:        "center",
		Animation:       "scale",
		Role:            "dialog",
	}
}
