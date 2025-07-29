import React, { useState } from 'react'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Upload, Images, Database, ExternalLink } from 'lucide-react'
import MediaUpload from '@/components/MediaUpload'
import MediaGallery from '@/components/MediaGallery'
import { isSupabaseConfigured, type MediaFile } from '@/lib/supabase'

const MediaManager = () => {
  const [refreshTrigger, setRefreshTrigger] = useState(0)

  const handleUploadSuccess = (files: MediaFile[]) => {
    // Trigger gallery refresh
    setRefreshTrigger(prev => prev + 1)
  }

  // Check if Supabase is configured
  if (!isSupabaseConfigured()) {
    return (
      <div className="min-h-screen bg-background p-6">
        <div className="max-w-4xl mx-auto">
          <div className="mb-8">
            <h1 className="text-3xl font-bold text-foreground mb-2">Media Manager</h1>
            <p className="text-muted-foreground">
              Upload and manage your media files with Supabase Storage and PostgreSQL
            </p>
          </div>

          <Card>
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                <Database className="w-5 h-5" />
                Supabase Connection Required
              </CardTitle>
            </CardHeader>
            <CardContent className="space-y-4">
              <p className="text-muted-foreground">
                To use the media upload functionality, you need to connect your project to Supabase. 
                This will enable:
              </p>
              
              <ul className="space-y-2 text-sm text-muted-foreground ml-6">
                <li className="flex items-center gap-2">
                  <span className="w-1.5 h-1.5 bg-primary rounded-full"></span>
                  PostgreSQL database for file metadata storage
                </li>
                <li className="flex items-center gap-2">
                  <span className="w-1.5 h-1.5 bg-primary rounded-full"></span>
                  Supabase Storage for secure file uploads
                </li>
                <li className="flex items-center gap-2">
                  <span className="w-1.5 h-1.5 bg-primary rounded-full"></span>
                  Authentication and authorization
                </li>
                <li className="flex items-center gap-2">
                  <span className="w-1.5 h-1.5 bg-primary rounded-full"></span>
                  Automatic file management and cleanup
                </li>
              </ul>

              <div className="pt-4">
                <p className="text-sm text-muted-foreground mb-4">
                  Click the green <strong>Supabase</strong> button in the top right corner of the Lovable interface to get started.
                </p>
                
                <Button variant="outline" asChild>
                  <a 
                    href="https://docs.lovable.dev/integrations/supabase/" 
                    target="_blank" 
                    rel="noopener noreferrer"
                    className="flex items-center gap-2"
                  >
                    <ExternalLink className="w-4 h-4" />
                    View Setup Guide
                  </a>
                </Button>
              </div>
            </CardContent>
          </Card>
        </div>
      </div>
    )
  }

  return (
    <div className="min-h-screen bg-background p-6">
      <div className="max-w-7xl mx-auto">
        <div className="mb-8">
          <h1 className="text-3xl font-bold text-foreground mb-2">Media Manager</h1>
          <p className="text-muted-foreground">
            Upload and manage your media files with Supabase Storage and PostgreSQL
          </p>
        </div>

        <Tabs defaultValue="upload" className="space-y-6">
          <TabsList className="grid w-[400px] grid-cols-2">
            <TabsTrigger value="upload" className="flex items-center space-x-2">
              <Upload className="w-4 h-4" />
              <span>Upload</span>
            </TabsTrigger>
            <TabsTrigger value="gallery" className="flex items-center space-x-2">
              <Images className="w-4 h-4" />
              <span>Gallery</span>
            </TabsTrigger>
          </TabsList>

          <TabsContent value="upload">
            <MediaUpload onUploadSuccess={handleUploadSuccess} />
          </TabsContent>

          <TabsContent value="gallery">
            <MediaGallery refreshTrigger={refreshTrigger} />
          </TabsContent>
        </Tabs>
      </div>
    </div>
  )
}

export default MediaManager