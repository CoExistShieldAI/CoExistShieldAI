import { Component, EventEmitter, Input, Output, inject, signal } from '@angular/core';
import { CommonModule } from '@angular/common';
import { AdminApiService } from '../../services/admin-api.service';

@Component({
  selector: 'app-image-upload',
  standalone: true,
  imports: [CommonModule],
  template: `
    <div class="space-y-3">
      <div class="flex items-center justify-between">
        <label class="block text-slate-300 font-medium text-xs">
          Photo Evidence / Field Attachments
        </label>
        <span class="text-[11px] font-mono" [class.text-emerald-400]="photoUrls.length < 5" [class.text-amber-400]="photoUrls.length >= 5">
          {{ photoUrls.length }} / {{ maxPhotos }} Photos
        </span>
      </div>

      <!-- Drag & Drop Zone -->
      @if (photoUrls.length < maxPhotos) {
        <div
          (dragover)="onDragOver($event)"
          (dragleave)="onDragLeave($event)"
          (drop)="onDrop($event)"
          (click)="fileInput.click()"
          [class.border-emerald-500]="isDragging()"
          [class.bg-emerald-950/20]="isDragging()"
          [class.border-slate-700]="!isDragging()"
          [class.bg-slate-900/60]="!isDragging()"
          class="border-2 border-dashed rounded-xl p-5 text-center cursor-pointer transition-all hover:border-slate-500 hover:bg-slate-900 flex flex-col items-center justify-center gap-2 group">
          
          <input
            #fileInput
            type="file"
            multiple
            accept="image/*"
            (change)="onFilesSelected($event)"
            class="hidden" />

          @if (isUploading()) {
            <div class="flex flex-col items-center gap-2 py-2">
              <svg class="animate-spin h-6 w-6 text-emerald-400" fill="none" viewBox="0 0 24 24">
                <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8v8H4z"></path>
              </svg>
              <span class="text-xs text-slate-300 font-medium">{{ uploadStatus() }}</span>
              <span class="text-[10px] text-slate-500">Auto-optimizing files &gt;1MB via client-side canvas</span>
            </div>
          } @else {
            <div class="w-10 h-10 rounded-full bg-slate-800 flex items-center justify-center text-slate-300 group-hover:bg-slate-700 group-hover:text-emerald-400 transition-colors">
              <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14m-6-6h.01M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z"></path>
              </svg>
            </div>
            <div>
              <p class="text-xs text-slate-300 font-medium">
                <span class="text-emerald-400 font-semibold underline decoration-emerald-500/40 underline-offset-2">Click to browse</span> or drag & drop images here
              </p>
              <p class="text-[10px] text-slate-500 mt-0.5">
                PNG, JPG, WEBP up to 5 photos. Files &gt;1MB are automatically compressed.
              </p>
            </div>
          }
        </div>
      }

      <!-- Alert notification if any -->
      @if (errorMessage()) {
        <div class="p-2.5 rounded-lg bg-rose-950/60 border border-rose-800 text-[11px] text-rose-300 flex items-center justify-between">
          <span>{{ errorMessage() }}</span>
          <button (click)="errorMessage.set('')" class="text-rose-400 hover:text-rose-200">✕</button>
        </div>
      }

      <!-- Uploaded Photos Grid -->
      @if (photoUrls.length > 0) {
        <div class="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-5 gap-3 pt-1">
          @for (url of photoUrls; track url; let idx = $index) {
            <div class="group relative rounded-xl overflow-hidden bg-slate-900 border border-slate-700 aspect-square shadow-md">
              <img [src]="url" [alt]="'Photo evidence ' + (idx + 1)" class="w-full h-full object-cover transition-transform duration-300 group-hover:scale-105" />
              
              <!-- Overlay with controls -->
              <div class="absolute inset-0 bg-gradient-to-t from-slate-950/80 via-transparent to-transparent opacity-0 group-hover:opacity-100 transition-opacity flex items-end justify-between p-2">
                <span class="text-[10px] font-mono text-slate-300 bg-slate-900/80 px-1.5 py-0.5 rounded">
                  #{{ idx + 1 }}
                </span>
                <div class="flex items-center gap-1">
                  <a [href]="url" target="_blank" rel="noopener noreferrer" class="p-1 rounded bg-slate-800/80 hover:bg-slate-700 text-slate-300 hover:text-white" title="View Full Image">
                    <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 6H6a2 2 0 00-2 2v10a2 2 0 002 2h10a2 2 0 002-2v-4M14 4h6m0 0v6m0-6L10 14"></path>
                    </svg>
                  </a>
                  <button type="button" (click)="removePhoto(idx)" class="p-1 rounded bg-rose-900/80 hover:bg-rose-800 text-rose-200 hover:text-white" title="Remove Photo">
                    <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"></path>
                    </svg>
                  </button>
                </div>
              </div>
            </div>
          }
        </div>
      }
    </div>
  `
})
export class ImageUploadComponent {
  private adminApi = inject(AdminApiService);

  @Input() photoUrls: string[] = [];
  @Output() photoUrlsChange = new EventEmitter<string[]>();

  readonly maxPhotos = 5;
  readonly isDragging = signal<boolean>(false);
  readonly isUploading = signal<boolean>(false);
  readonly uploadStatus = signal<string>('');
  readonly errorMessage = signal<string>('');

  onDragOver(event: DragEvent) {
    event.preventDefault();
    event.stopPropagation();
    this.isDragging.set(true);
  }

  onDragLeave(event: DragEvent) {
    event.preventDefault();
    event.stopPropagation();
    this.isDragging.set(false);
  }

  onDrop(event: DragEvent) {
    event.preventDefault();
    event.stopPropagation();
    this.isDragging.set(false);

    if (event.dataTransfer && event.dataTransfer.files) {
      this.processFiles(Array.from(event.dataTransfer.files));
    }
  }

  onFilesSelected(event: Event) {
    const input = event.target as HTMLInputElement;
    if (input.files && input.files.length > 0) {
      this.processFiles(Array.from(input.files));
      input.value = ''; // Reset input to allow re-selection
    }
  }

  removePhoto(index: number) {
    const updated = [...this.photoUrls];
    updated.splice(index, 1);
    this.photoUrls = updated;
    this.photoUrlsChange.emit(this.photoUrls);
  }

  private async processFiles(files: File[]) {
    this.errorMessage.set('');

    // Filter only image types
    const imageFiles = files.filter(f => f.type.startsWith('image/'));
    if (imageFiles.length === 0) {
      this.errorMessage.set('Please select valid image files (JPEG, PNG, WEBP).');
      return;
    }

    const availableSlots = this.maxPhotos - this.photoUrls.length;
    if (availableSlots <= 0) {
      this.errorMessage.set(`Maximum of ${this.maxPhotos} photos reached.`);
      return;
    }

    let toProcess = imageFiles;
    if (toProcess.length > availableSlots) {
      this.errorMessage.set(`Only ${availableSlots} more photo(s) can be added (max ${this.maxPhotos}).`);
      toProcess = toProcess.slice(0, availableSlots);
    }

    this.isUploading.set(true);
    this.uploadStatus.set(`Compressing & preparing ${toProcess.length} image(s)...`);

    try {
      // Client-side shrink for any files exceeding 1MB
      const compressedFiles: File[] = [];
      let compressedCount = 0;

      for (let i = 0; i < toProcess.length; i++) {
        const file = toProcess[i];
        if (file.size > 1024 * 1024) {
          this.uploadStatus.set(`Optimizing image ${i + 1}/${toProcess.length} (<1MB)...`);
          const compressed = await this.compressImage(file);
          compressedFiles.push(compressed);
          compressedCount++;
        } else {
          compressedFiles.push(file);
        }
      }

      this.uploadStatus.set(`Uploading to Cloud Storage...`);
      this.adminApi.uploadImages(compressedFiles).subscribe({
        next: (res) => {
          this.isUploading.set(false);
          const newUrls = [...this.photoUrls, ...res.urls].slice(0, this.maxPhotos);
          this.photoUrls = newUrls;
          this.photoUrlsChange.emit(this.photoUrls);
          if (compressedCount > 0) {
            this.uploadStatus.set(`Optimized ${compressedCount} photo(s) to under 1MB.`);
          }
        },
        error: (err) => {
          this.isUploading.set(false);
          this.errorMessage.set('Failed to upload photos to Cloud Storage. Please try again.');
          console.error('Image upload error:', err);
        }
      });
    } catch (err) {
      this.isUploading.set(false);
      this.errorMessage.set('An error occurred while optimizing images.');
      console.error(err);
    }
  }

  // Client-side shrink using HTML5 Canvas
  private compressImage(file: File): Promise<File> {
    const ONE_MB = 1024 * 1024;
    return new Promise((resolve) => {
      const img = new Image();
      const objectUrl = URL.createObjectURL(file);

      img.onload = () => {
        URL.revokeObjectURL(objectUrl);
        let width = img.width;
        let height = img.height;
        const maxDimension = 1920;

        // Proportional dimension clamping
        if (width > maxDimension || height > maxDimension) {
          if (width > height) {
            height = Math.round((height * maxDimension) / width);
            width = maxDimension;
          } else {
            width = Math.round((width * maxDimension) / height);
            height = maxDimension;
          }
        }

        const canvas = document.createElement('canvas');
        canvas.width = width;
        canvas.height = height;
        const ctx = canvas.getContext('2d');
        if (!ctx) {
          resolve(file);
          return;
        }

        ctx.drawImage(img, 0, 0, width, height);

        // Quality stepping: 0.82 -> 0.70 -> 0.55 if still > 1MB
        const tryExport = (quality: number) => {
          canvas.toBlob((blob) => {
            if (!blob) {
              resolve(file);
              return;
            }
            if (blob.size <= ONE_MB || quality <= 0.55) {
              const cleanName = file.name.replace(/\.[^/.]+$/, '') + '.jpg';
              const compressedFile = new File([blob], cleanName, { type: 'image/jpeg' });
              resolve(compressedFile);
            } else {
              tryExport(quality - 0.15);
            }
          }, 'image/jpeg', quality);
        };

        tryExport(0.82);
      };

      img.onerror = () => {
        URL.revokeObjectURL(objectUrl);
        resolve(file);
      };

      img.src = objectUrl;
    });
  }
}
