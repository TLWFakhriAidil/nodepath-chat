import React from 'react'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Database, ExternalLink } from 'lucide-react'

const MediaManager = () => {
  return (
    <div className="min-h-screen bg-background p-6">
      <div className="max-w-4xl mx-auto">
        <div className="mb-8">
          <h1 className="text-3xl font-bold text-foreground mb-2">Media Manager</h1>
          <p className="text-muted-foreground">
            Media management functionality has been removed along with Supabase
          </p>
        </div>

        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Database className="w-5 h-5" />
              Media Features Removed
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            <p className="text-muted-foreground">
              The following features have been removed as part of Supabase disconnection:
            </p>
            
            <ul className="space-y-2 text-sm text-muted-foreground ml-6">
              <li className="flex items-center gap-2">
                <span className="w-1.5 h-1.5 bg-destructive rounded-full"></span>
                File upload functionality
              </li>
              <li className="flex items-center gap-2">
                <span className="w-1.5 h-1.5 bg-destructive rounded-full"></span>
                Media gallery browser
              </li>
              <li className="flex items-center gap-2">
                <span className="w-1.5 h-1.5 bg-destructive rounded-full"></span>
                File storage management
              </li>
              <li className="flex items-center gap-2">
                <span className="w-1.5 h-1.5 bg-destructive rounded-full"></span>
                Media picker components
              </li>
            </ul>

            <div className="pt-4">
              <p className="text-sm text-muted-foreground mb-4">
                To restore media functionality, you would need to implement a direct file storage solution 
                or alternative cloud storage service.
              </p>
              
              <Button variant="outline" asChild>
                <a 
                  href="#" 
                  className="flex items-center gap-2 pointer-events-none opacity-50"
                >
                  <ExternalLink className="w-4 h-4" />
                  Media Features Disabled
                </a>
              </Button>
            </div>
          </CardContent>
        </Card>
      </div>
    </div>
  )
}

export default MediaManager