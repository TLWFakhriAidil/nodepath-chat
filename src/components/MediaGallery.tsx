import React, { useState, useEffect } from 'react'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Input } from '@/components/ui/input'
import { AlertDialog, AlertDialogAction, AlertDialogCancel, AlertDialogContent, AlertDialogDescription, AlertDialogFooter, AlertDialogHeader, AlertDialogTitle, AlertDialogTrigger } from '@/components/ui/alert-dialog'
import { Image, Video, Music, File, Trash2, Search, Download, Copy, Check } from 'lucide-react'
import { supabase, isSupabaseConfigured, type MediaFile } from '@/lib/supabase'
import { useToast } from '@/hooks/use-toast'
import { formatDistanceToNow } from 'date-fns'

interface MediaGalleryProps {
  refreshTrigger?: number
}

const MediaGallery: React.FC<MediaGalleryProps> = ({ refreshTrigger }) => {
  const [mediaFiles, setMediaFiles] = useState<MediaFile[]>([])
  const [loading, setLoading] = useState(true)
  const [searchTerm, setSearchTerm] = useState('')
  const [copiedUrl, setCopiedUrl] = useState<string | null>(null)
  const { toast } = useToast()

  const fetchMediaFiles = async () => {
    if (!supabase) {
      setLoading(false)
      return
    }
    
    try {
      const { data, error } = await supabase
        .from('media_files')
        .select('*')
        .order('uploaded_at', { ascending: false })

      if (error) {
        throw error
      }

      setMediaFiles(data || [])
    } catch (error) {
      toast({
        title: "Error loading files",
        description: error instanceof Error ? error.message : 'Unknown error',
        variant: "destructive"
      })
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    fetchMediaFiles()
  }, [refreshTrigger])

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

  const handleDelete = async (file: MediaFile) => {
    if (!supabase) return
    
    try {
      // Delete from storage
      const { error: storageError } = await supabase.storage
        .from('media')
        .remove([file.storage_path])

      if (storageError) {
        throw storageError
      }

      // Delete from database
      const { error: dbError } = await supabase
        .from('media_files')
        .delete()
        .eq('id', file.id)

      if (dbError) {
        throw dbError
      }

      setMediaFiles(prev => prev.filter(f => f.id !== file.id))
      
      toast({
        title: "File deleted",
        description: `${file.original_name} has been deleted successfully`
      })
    } catch (error) {
      toast({
        title: "Error deleting file",
        description: error instanceof Error ? error.message : 'Unknown error',
        variant: "destructive"
      })
    }
  }

  const copyToClipboard = async (url: string) => {
    try {
      await navigator.clipboard.writeText(url)
      setCopiedUrl(url)
      setTimeout(() => setCopiedUrl(null), 2000)
      
      toast({
        title: "URL copied",
        description: "File URL has been copied to clipboard"
      })
    } catch (error) {
      toast({
        title: "Failed to copy",
        description: "Could not copy URL to clipboard",
        variant: "destructive"
      })
    }
  }

  const filteredFiles = mediaFiles.filter(file =>
    file.original_name.toLowerCase().includes(searchTerm.toLowerCase()) ||
    file.file_type.toLowerCase().includes(searchTerm.toLowerCase())
  )

  const renderMediaPreview = (file: MediaFile) => {
    if (file.file_type.startsWith('image/')) {
      return (
        <img
          src={file.public_url}
          alt={file.original_name}
          className="w-full h-32 object-cover rounded-md"
        />
      )
    }
    
    if (file.file_type.startsWith('video/')) {
      return (
        <video
          src={file.public_url}
          className="w-full h-32 object-cover rounded-md"
          controls={false}
        />
      )
    }
    
    if (file.file_type.startsWith('audio/')) {
      return (
        <div className="w-full h-32 bg-muted rounded-md flex items-center justify-center">
          <Music className="w-12 h-12 text-muted-foreground" />
        </div>
      )
    }

    return (
      <div className="w-full h-32 bg-muted rounded-md flex items-center justify-center">
        <File className="w-12 h-12 text-muted-foreground" />
      </div>
    )
  }

  if (loading) {
    return (
      <Card>
        <CardContent className="p-6">
          <div className="flex items-center justify-center h-32">
            <div className="text-muted-foreground">Loading media files...</div>
          </div>
        </CardContent>
      </Card>
    )
  }

  return (
    <Card>
      <CardHeader>
        <div className="flex items-center justify-between">
          <CardTitle className="flex items-center gap-2">
            <File className="w-5 h-5" />
            Media Gallery ({mediaFiles.length} files)
          </CardTitle>
          <div className="flex items-center gap-2">
            <div className="relative">
              <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 w-4 h-4 text-muted-foreground" />
              <Input
                placeholder="Search files..."
                value={searchTerm}
                onChange={(e) => setSearchTerm(e.target.value)}
                className="pl-10 w-64"
              />
            </div>
          </div>
        </div>
      </CardHeader>
      <CardContent>
        {filteredFiles.length === 0 ? (
          <div className="text-center py-8 text-muted-foreground">
            {searchTerm ? 'No files match your search.' : 'No media files uploaded yet.'}
          </div>
        ) : (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-4">
            {filteredFiles.map((file) => (
              <Card key={file.id} className="overflow-hidden">
                <div className="relative">
                  {renderMediaPreview(file)}
                  <div className="absolute top-2 right-2">
                    <Badge className={getFileTypeColor(file.file_type)} variant="secondary">
                      {file.file_type.split('/')[0]}
                    </Badge>
                  </div>
                </div>
                
                <CardContent className="p-4">
                  <div className="space-y-2">
                    <div className="flex items-start gap-2">
                      {getFileIcon(file.file_type)}
                      <div className="flex-1 min-w-0">
                        <p className="font-medium text-sm truncate" title={file.original_name}>
                          {file.original_name}
                        </p>
                        <p className="text-xs text-muted-foreground">
                          {(file.file_size / 1024 / 1024).toFixed(2)} MB
                        </p>
                        <p className="text-xs text-muted-foreground">
                          {formatDistanceToNow(new Date(file.uploaded_at), { addSuffix: true })}
                        </p>
                      </div>
                    </div>
                    
                    <div className="flex items-center gap-1">
                      <Button
                        variant="ghost"
                        size="sm"
                        onClick={() => copyToClipboard(file.public_url)}
                        className="flex-1"
                      >
                        {copiedUrl === file.public_url ? (
                          <Check className="w-4 h-4" />
                        ) : (
                          <Copy className="w-4 h-4" />
                        )}
                      </Button>
                      
                      <Button
                        variant="ghost"
                        size="sm"
                        asChild
                        className="flex-1"
                      >
                        <a
                          href={file.public_url}
                          download={file.original_name}
                          target="_blank"
                          rel="noopener noreferrer"
                        >
                          <Download className="w-4 h-4" />
                        </a>
                      </Button>
                      
                      <AlertDialog>
                        <AlertDialogTrigger asChild>
                          <Button variant="ghost" size="sm" className="flex-1 text-destructive hover:text-destructive">
                            <Trash2 className="w-4 h-4" />
                          </Button>
                        </AlertDialogTrigger>
                        <AlertDialogContent>
                          <AlertDialogHeader>
                            <AlertDialogTitle>Delete File</AlertDialogTitle>
                            <AlertDialogDescription>
                              Are you sure you want to delete "{file.original_name}"? This action cannot be undone.
                            </AlertDialogDescription>
                          </AlertDialogHeader>
                          <AlertDialogFooter>
                            <AlertDialogCancel>Cancel</AlertDialogCancel>
                            <AlertDialogAction
                              onClick={() => handleDelete(file)}
                              className="bg-destructive text-destructive-foreground hover:bg-destructive/90"
                            >
                              Delete
                            </AlertDialogAction>
                          </AlertDialogFooter>
                        </AlertDialogContent>
                      </AlertDialog>
                    </div>
                  </div>
                </CardContent>
              </Card>
            ))}
          </div>
        )}
      </CardContent>
    </Card>
  )
}

export default MediaGallery