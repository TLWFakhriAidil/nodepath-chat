import React, { useState, useCallback } from 'react'
import { useDropzone } from 'react-dropzone'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Progress } from '@/components/ui/progress'
import { Badge } from '@/components/ui/badge'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { Upload, File, Image, Video, Music, X, Check } from 'lucide-react'
import { supabase, type MediaFile } from '@/lib/supabase'
import { useToast } from '@/hooks/use-toast'

const ACCEPTED_FILE_TYPES = {
  'image/*': ['.png', '.jpg', '.jpeg'],
  'video/mp4': ['.mp4'],
  'audio/*': ['.mp3', '.wav']
}

const MAX_FILE_SIZE = 25 * 1024 * 1024 // 25MB

interface UploadedFile {
  file: File
  id: string
  progress: number
  status: 'uploading' | 'success' | 'error'
  url?: string
  mediaFile?: MediaFile
}

interface MediaUploadProps {
  onUploadSuccess?: (files: MediaFile[]) => void
}

const MediaUpload: React.FC<MediaUploadProps> = ({ onUploadSuccess }) => {
  const [uploadedFiles, setUploadedFiles] = useState<UploadedFile[]>([])
  const { toast } = useToast()

  const getFileIcon = (type: string) => {
    if (type.startsWith('image/')) return <Image className="w-4 h-4" />
    if (type.startsWith('video/')) return <Video className="w-4 h-4" />
    if (type.startsWith('audio/')) return <Music className="w-4 h-4" />
    return <File className="w-4 h-4" />
  }

  const getFileTypeColor = (type: string) => {
    if (type.startsWith('image/')) return 'bg-green-100 text-green-800'
    if (type.startsWith('video/')) return 'bg-blue-100 text-blue-800'
    if (type.startsWith('audio/')) return 'bg-purple-100 text-purple-800'
    return 'bg-gray-100 text-gray-800'
  }

  const uploadFile = async (file: File): Promise<MediaFile> => {
    const fileExt = file.name.split('.').pop()
    const fileName = `${Date.now()}-${Math.random().toString(36).substring(2)}.${fileExt}`
    const filePath = `uploads/${fileName}`

    // Upload to Supabase Storage
    const { data: uploadData, error: uploadError } = await supabase.storage
      .from('media')
      .upload(filePath, file)

    if (uploadError) {
      throw new Error(`Upload failed: ${uploadError.message}`)
    }

    // Get public URL
    const { data: { publicUrl } } = supabase.storage
      .from('media')
      .getPublicUrl(filePath)

    // Save metadata to database
    const { data: mediaFile, error: dbError } = await supabase
      .from('media_files')
      .insert({
        filename: fileName,
        original_name: file.name,
        file_type: file.type,
        file_size: file.size,
        storage_path: filePath,
        public_url: publicUrl
      })
      .select()
      .single()

    if (dbError) {
      // Clean up uploaded file if database insert fails
      await supabase.storage.from('media').remove([filePath])
      throw new Error(`Database error: ${dbError.message}`)
    }

    return mediaFile
  }

  const handleUpload = useCallback(async (acceptedFiles: File[]) => {
    const newFiles: UploadedFile[] = acceptedFiles.map(file => ({
      file,
      id: Math.random().toString(36),
      progress: 0,
      status: 'uploading' as const
    }))

    setUploadedFiles(prev => [...prev, ...newFiles])

    const uploadPromises = newFiles.map(async (uploadedFile) => {
      try {
        setUploadedFiles(prev => 
          prev.map(f => f.id === uploadedFile.id ? { ...f, progress: 50 } : f)
        )

        const mediaFile = await uploadFile(uploadedFile.file)

        setUploadedFiles(prev => 
          prev.map(f => f.id === uploadedFile.id ? { 
            ...f, 
            progress: 100, 
            status: 'success',
            url: mediaFile.public_url,
            mediaFile
          } : f)
        )

        return mediaFile
      } catch (error) {
        setUploadedFiles(prev => 
          prev.map(f => f.id === uploadedFile.id ? { ...f, status: 'error' } : f)
        )
        
        toast({
          title: "Upload failed",
          description: `Failed to upload ${uploadedFile.file.name}: ${error instanceof Error ? error.message : 'Unknown error'}`,
          variant: "destructive"
        })
        
        return null
      }
    })

    const results = await Promise.all(uploadPromises)
    const successfulUploads = results.filter((result): result is MediaFile => result !== null)
    
    if (successfulUploads.length > 0) {
      onUploadSuccess?.(successfulUploads)
      toast({
        title: "Upload successful",
        description: `${successfulUploads.length} file(s) uploaded successfully`
      })
    }
  }, [onUploadSuccess, toast])

  const { getRootProps, getInputProps, isDragActive } = useDropzone({
    onDrop: handleUpload,
    accept: ACCEPTED_FILE_TYPES,
    maxSize: MAX_FILE_SIZE,
    onDropRejected: (rejectedFiles) => {
      rejectedFiles.forEach(({ file, errors }) => {
        errors.forEach(error => {
          if (error.code === 'file-too-large') {
            toast({
              title: "File too large",
              description: `${file.name} is larger than 25MB`,
              variant: "destructive"
            })
          } else if (error.code === 'file-invalid-type') {
            toast({
              title: "Invalid file type",
              description: `${file.name} is not a supported file type`,
              variant: "destructive"
            })
          }
        })
      })
    }
  })

  const removeFile = (id: string) => {
    setUploadedFiles(prev => prev.filter(f => f.id !== id))
  }

  const clearAll = () => {
    setUploadedFiles([])
  }

  return (
    <div className="space-y-6">
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Upload className="w-5 h-5" />
            Media Upload
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div
            {...getRootProps()}
            className={`border-2 border-dashed rounded-lg p-8 text-center cursor-pointer transition-colors ${
              isDragActive 
                ? 'border-primary bg-primary/5' 
                : 'border-muted-foreground/25 hover:border-primary/50'
            }`}
          >
            <input {...getInputProps()} />
            <Upload className="w-12 h-12 mx-auto mb-4 text-muted-foreground" />
            {isDragActive ? (
              <p className="text-lg font-medium">Drop files here...</p>
            ) : (
              <div>
                <p className="text-lg font-medium mb-2">
                  Drag & drop files here, or click to select
                </p>
                <p className="text-sm text-muted-foreground mb-4">
                  Supports images (PNG, JPG), videos (MP4), and audio (MP3, WAV)
                </p>
                <p className="text-xs text-muted-foreground">
                  Maximum file size: 25MB
                </p>
              </div>
            )}
          </div>

          <Alert className="mt-4">
            <AlertDescription>
              Files will be stored securely in Supabase Storage with metadata saved to your PostgreSQL database.
            </AlertDescription>
          </Alert>
        </CardContent>
      </Card>

      {uploadedFiles.length > 0 && (
        <Card>
          <CardHeader className="flex flex-row items-center justify-between">
            <CardTitle>Upload Progress</CardTitle>
            <Button variant="outline" size="sm" onClick={clearAll}>
              Clear All
            </Button>
          </CardHeader>
          <CardContent className="space-y-4">
            {uploadedFiles.map((uploadedFile) => (
              <div key={uploadedFile.id} className="flex items-center gap-4 p-4 border rounded-lg">
                <div className="flex items-center gap-2 flex-1 min-w-0">
                  {getFileIcon(uploadedFile.file.type)}
                  <div className="flex-1 min-w-0">
                    <p className="font-medium truncate">{uploadedFile.file.name}</p>
                    <p className="text-sm text-muted-foreground">
                      {(uploadedFile.file.size / 1024 / 1024).toFixed(2)} MB
                    </p>
                  </div>
                  <Badge className={getFileTypeColor(uploadedFile.file.type)}>
                    {uploadedFile.file.type.split('/')[0]}
                  </Badge>
                </div>

                <div className="flex items-center gap-2">
                  {uploadedFile.status === 'uploading' && (
                    <Progress value={uploadedFile.progress} className="w-20" />
                  )}
                  {uploadedFile.status === 'success' && (
                    <Check className="w-5 h-5 text-green-600" />
                  )}
                  {uploadedFile.status === 'error' && (
                    <X className="w-5 h-5 text-red-600" />
                  )}
                  <Button
                    variant="ghost"
                    size="sm"
                    onClick={() => removeFile(uploadedFile.id)}
                  >
                    <X className="w-4 h-4" />
                  </Button>
                </div>
              </div>
            ))}
          </CardContent>
        </Card>
      )}
    </div>
  )
}

export default MediaUpload