import { useCallback } from 'react'
import { Upload } from 'lucide-react'
import { cn } from '@/lib/utils'

interface UploadZoneProps {
  onUpload: (file: File) => void
  uploading?: boolean
  progress?: number
}

export function UploadZone({ onUpload, uploading, progress }: UploadZoneProps) {
  const handleDrop = useCallback(
    (e: React.DragEvent) => {
      e.preventDefault()
      const file = e.dataTransfer.files[0]
      if (file && file.name.endsWith('.csv')) onUpload(file)
    },
    [onUpload]
  )

  const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0]
    if (file) onUpload(file)
  }

  return (
    <div
      onDragOver={(e) => e.preventDefault()}
      onDrop={handleDrop}
      onClick={() => document.getElementById('csv-upload')?.click()}
      className={cn(
        'border-2 border-dashed border-default rounded-2xl p-10 lg:p-12 text-center cursor-pointer transition-all duration-200',
        'hover:border-noant-sky hover:bg-noant-sky/5'
      )}
    >
      <div className="w-14 h-14 lg:w-16 lg:h-16 bg-inset text-noant-sky rounded-2xl flex items-center justify-center mx-auto mb-4">
        <Upload className="w-6 h-6 lg:w-7 lg:h-7" />
      </div>
      <p className="text-base font-semibold text-primary mb-1">Drop your CSV here</p>
      <p className="text-sm text-secondary mb-5">Format: Category, Question, Answer</p>
      <input id="csv-upload" type="file" accept=".csv" onChange={handleChange} className="hidden" />
      <span className="inline-flex px-5 py-2.5 bg-noant-sky text-white rounded-xl text-sm font-semibold hover:bg-noant-sky-deep transition-colors btn-press">
        Choose file
      </span>

      {uploading && (
        <div className="mt-6 max-w-xs mx-auto">
          <div className="h-1.5 bg-inset rounded-full overflow-hidden">
            <div
              className="h-full bg-noant-sky rounded-full transition-all duration-300"
              style={{ width: `${progress}%` }}
            />
          </div>
          <p className="text-xs text-secondary mt-2">{progress}% uploaded</p>
        </div>
      )}
    </div>
  )
}
